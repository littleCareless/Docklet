#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import subprocess
import socket
import re
import os
import paramiko  # pip install paramiko

# 这里配置你的 OpenWrt 账号密码
ROUTER_USER = "root"
ROUTER_PASSWORD = "password"  # 请替换成实际密码

def get_open_ports():
    """
    使用 ss 或 netstat 获取本机的监听端口，自动去重
    """
    ports_set = set()  # 使用set自动去重
    raw_ports = []  # 用于调试，记录去重前的端口
    
    try:
        result = subprocess.check_output(["ss", "-tuln"]).decode()
    except FileNotFoundError:
        result = subprocess.check_output(["netstat", "-tuln"]).decode()

    for line in result.splitlines():
        if not line or re.match(r'^Netid|^Proto', line):
            continue

        columns = line.split()
        if len(columns) < 5:
            continue

        proto = columns[0].lower()
        local_address = columns[4]

        if proto not in ["tcp", "udp"]:
            continue

        if ':' in local_address:
            port = local_address.split(':')[-1]
            try:
                port = int(port)
            except ValueError:
                continue
            raw_ports.append((proto, port))  # 记录原始端口（包含重复）
            ports_set.add((proto, port))  # 添加到set中自动去重

    # 转换为排序的列表
    ports_list = sorted(list(ports_set), key=lambda x: (x[0], x[1]))
    
    # 调试输出（仅在调试模式下显示）
    debug_mode = os.getenv('DEBUG', '').lower() in ['1', 'true', 'yes']
    if debug_mode:
        print(f"\n[调试] 端口去重统计: 原始 {len(raw_ports)} 个，去重后 {len(ports_list)} 个")
        if len(raw_ports) != len(ports_list):
            print(f"[调试] 发现并移除了 {len(raw_ports) - len(ports_list)} 个重复端口")
    
    return ports_list

def get_local_ip():
    """
    获取本机的 IP 地址
    """
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        local_ip = s.getsockname()[0]
        s.close()
        return local_ip
    except Exception:
        return "无法获取"

def get_default_gateway():
    """
    获取默认网关（OpenWrt 路由器的 IP）
    """
    try:
        result = subprocess.check_output(["ip", "route", "show", "default"]).decode()
        for line in result.splitlines():
            if line.startswith("default"):
                gateway_ip = line.split()[2]
                return gateway_ip
        return "无法获取"
    except Exception:
        return "无法获取"

def parse_uci_redirect_rules(uci_output):
    """
    解析UCI重定向规则输出，返回规则列表
    """
    rules_by_index = {}
    
    for line in uci_output.splitlines():
        # 格式类似 firewall.@redirect[0].src='wan'
        m = re.match(r"firewall\.@redirect\[(\d+)\]\.(\w+)='?([^']+)'?", line)
        if m:
            idx, key, value = m.groups()
            idx = int(idx)
            if idx not in rules_by_index:
                rules_by_index[idx] = {}
            rules_by_index[idx][key] = value
    
    # 将按索引分组的规则转换为列表
    existing_rules = []
    for idx in sorted(rules_by_index.keys()):
        existing_rules.append(rules_by_index[idx])
    
    return existing_rules

def rule_exists(existing_rules, proto, src_dport, dest_ip, dest_port):
    """
    检查指定的端口转发规则是否已存在
    """
    for rule in existing_rules:
        if (rule.get("proto") == proto and
            rule.get("src_dport") == str(src_dport) and
            rule.get("dest_ip") == dest_ip and
            rule.get("dest_port") == str(dest_port)):
            return True
    return False

def update_openwrt_firewall(router_ip, local_ip, ports, debug=False):
    """
    通过 SSH 登录 OpenWrt 并更新防火墙转发规则，避免重复添加
    """
    try:
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(router_ip, username=ROUTER_USER, password=ROUTER_PASSWORD)

        # 1. 获取所有现有的 redirect 规则
        stdin, stdout, stderr = ssh.exec_command("uci show firewall | grep redirect")
        existing_rules_raw = stdout.read().decode()
        
        if debug:
            print("当前防火墙重定向规则:\n", existing_rules_raw)

        # 2. 解析现有规则
        existing_rules = parse_uci_redirect_rules(existing_rules_raw)
        
        if debug:
            print("\n解析后的现有规则:")
            for i, rule in enumerate(existing_rules):
                print(f"规则 {i}: {rule}")

        # 3. 遍历待添加端口，检查是否已存在
        added_count = 0
        skipped_count = 0
        
        for proto, port in ports:
            if rule_exists(existing_rules, proto, port, local_ip, port):
                print(f"规则已存在，跳过添加: 协议 {proto} 端口 {port} 转发至 {local_ip}:{port}")
                skipped_count += 1
                continue

            name = f"auto-{proto}-{port}"
            uci_commands = f"""
uci add firewall redirect
uci set firewall.@redirect[-1].name='{name}'
uci set firewall.@redirect[-1].target='DNAT'
uci set firewall.@redirect[-1].src='wan'
uci set firewall.@redirect[-1].dest='lan'
uci set firewall.@redirect[-1].proto='{proto}'
uci set firewall.@redirect[-1].src_dport='{port}'
uci set firewall.@redirect[-1].dest_ip='{local_ip}'
uci set firewall.@redirect[-1].dest_port='{port}'
"""
            print(f"\n添加防火墙转发规则：{name} 协议: {proto} 端口: {port} 转发至 {local_ip}:{port}")

            # 执行UCI命令
            stdin, stdout, stderr = ssh.exec_command(uci_commands)
            # 确保命令执行完成
            stdout.channel.recv_exit_status()
            added_count += 1

        # 4. 提交并重载防火墙（仅在有新增规则时）
        if added_count > 0:
            ssh.exec_command("uci commit firewall && /etc/init.d/firewall reload")
            print(f"\n防火墙规则已更新并重载完成。新增 {added_count} 条规则，跳过 {skipped_count} 条已存在规则。")
        else:
            print(f"\n无需更新防火墙规则。跳过 {skipped_count} 条已存在规则。")

        ssh.close()

    except Exception as e:
        print(f"\n[错误] SSH 更新失败: {e}")

def main():
    local_ip = get_local_ip()
    router_ip = get_default_gateway()
    ports = get_open_ports()

    print(f"\n本机 IP 地址：{local_ip}")
    print(f"OpenWrt 路由器（默认网关）IP：{router_ip}")

    if not ports:
        print("\n当前主机没有监听的端口。")
        return

    print("\n当前主机监听的端口：\n")
    for proto, port in sorted(ports, key=lambda x: (x[0], x[1])):
        print(f"协议: {proto.upper()}  端口: {port}")

    # 默认直接执行 SSH 更新，可以通过参数启用调试模式
    print("\n正在自动将端口同步到 OpenWrt 防火墙...")
    debug_mode = os.getenv('DEBUG', '').lower() in ['1', 'true', 'yes']
    update_openwrt_firewall(router_ip, local_ip, ports, debug=debug_mode)

if __name__ == "__main__":
    main()