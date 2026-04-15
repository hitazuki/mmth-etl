#!/usr/bin/env python3
"""
按日期截取日志文件
用法: python3 extract_by_date.py <源日志文件> <日期>
示例: python3 extract_by_date.py ./test-json.log 2026-04-12
输出: test-2026-04-12.log
"""

import sys
import json
from datetime import datetime


def extract_by_date(source_file: str, target_date: str):
    """从源日志文件中提取指定日期的记录"""
    output_file = f"test-{target_date}.log"

    # 验证日期格式
    try:
        datetime.strptime(target_date, "%Y-%m-%d")
    except ValueError:
        print(f"错误: 日期格式应为 YYYY-MM-DD，例如: 2026-04-12")
        sys.exit(1)

    count = 0
    total = 0

    with open(source_file, 'r', encoding='utf-8') as f_in, \
         open(output_file, 'w', encoding='utf-8') as f_out:

        for line_num, line in enumerate(f_in, 1):
            total += 1
            line = line.strip()
            if not line:
                continue

            try:
                entry = json.loads(line)
                # 从 time 字段提取日期 (UTC格式: 2026-04-12T15:04:05Z)
                time_str = entry.get('time', '')
                if time_str:
                    # 提取日期部分
                    entry_date = time_str[:10]
                    if entry_date == target_date:
                        f_out.write(line + '\n')
                        count += 1
            except json.JSONDecodeError:
                print(f"警告: 第 {line_num} 行 JSON 解析失败，跳过")
                continue

    print(f"完成: 从 {total} 行中提取了 {count} 行 {target_date} 的记录")
    print(f"输出文件: {output_file}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("用法: python3 extract_by_date.py <源日志文件> <日期>")
        print("示例: python3 extract_by_date.py ./test-json.log 2026-04-12")
        sys.exit(1)

    source_file = sys.argv[1]
    target_date = sys.argv[2]

    extract_by_date(source_file, target_date)
