#!/usr/bin/env python3
"""
Excel to CSV Migration Script
将现有的剧集Excel数据导出为CSV格式,用于导入到Go系统
"""

import pandas as pd
import os
import sys
from datetime import datetime

# 配置
EXCEL_FILE = "../../../py/data/剧集列表.xlsx"
OUTPUT_DIR = "./output"

# 创建输出目录
os.makedirs(OUTPUT_DIR, exist_ok=True)


def export_shows_to_csv():
    """导出剧集数据到CSV"""
    print("正在读取剧集Excel文件...")
    
    try:
        # 读取Excel文件
        df = pd.read_excel(EXCEL_FILE)
        
        print(f"找到 {len(df)} 条剧集记录")
        print(f"列名: {df.columns.tolist()}")
        
        # 检查必要的列
        required_columns = ['TMDB_ID', '名称']
        missing_columns = [col for col in required_columns if col not in df.columns]
        
        if missing_columns:
            print(f"错误: Excel文件缺少必要的列: {missing_columns}")
            print(f"可用的列: {df.columns.tolist()}")
            sys.exit(1)
        
        # 清理数据
        df = df.fillna('')
        
        # 导出为CSV
        output_file = os.path.join(OUTPUT_DIR, "shows.csv")
        df.to_csv(output_file, index=False, encoding='utf-8-sig')
        
        print(f"✅ 剧集数据已导出到: {output_file}")
        return True
        
    except FileNotFoundError:
        print(f"错误: 找不到文件 {EXCEL_FILE}")
        print("请确保Excel文件存在于正确位置")
        sys.exit(1)
    except Exception as e:
        print(f"错误: {str(e)}")
        sys.exit(1)


def export_episodes_to_csv():
    """导出剧集详情到CSV"""
    print("\n正在读取剧集详情...")
    
    episodes_dir = "../../../py/data/剧集详情"
    
    if not os.path.exists(episodes_dir):
        print(f"警告: 剧集详情目录不存在: {episodes_dir}")
        return False
    
    all_episodes = []
    
    # 遍历所有剧集详情文件
    for filename in os.listdir(episodes_dir):
        if not filename.endswith('.xlsx'):
            continue
        
        filepath = os.path.join(episodes_dir, filename)
        print(f"处理文件: {filename}")
        
        try:
            # 从文件名提取TMDB ID
            # 格式: 剧集名（年份）-TMDB_ID.xlsx 或 剧集名-TMDB_ID.xlsx
            tmdb_id = None
            if '-' in filename:
                tmdb_id = filename.split('-')[-1].replace('.xlsx', '')
            
            if not tmdb_id or not tmdb_id.isdigit():
                print(f"  跳过: 无法提取TMDB ID")
                continue
            
            # 读取Excel文件
            df = pd.read_excel(filepath)
            
            # 添加TMDB ID列
            df['TMDB_ID'] = int(tmdb_id)
            
            all_episodes.append(df)
            
        except Exception as e:
            print(f"  错误: {str(e)}")
            continue
    
    if not all_episodes:
        print("警告: 没有找到有效的剧集详情数据")
        return False
    
    # 合并所有数据
    episodes_df = pd.concat(all_episodes, ignore_index=True)
    
    # 清理数据
    episodes_df = episodes_df.fillna('')
    
    # 导出为CSV
    output_file = os.path.join(OUTPUT_DIR, "episodes.csv")
    episodes_df.to_csv(output_file, index=False, encoding='utf-8-sig')
    
    print(f"✅ 剧集详情已导出到: {output_file}")
    print(f"总共 {len(episodes_df)} 条剧集记录")
    return True


def generate_migration_report():
    """生成迁移报告"""
    report_file = os.path.join(OUTPUT_DIR, "migration_report.txt")
    
    with open(report_file, 'w', encoding='utf-8') as f:
        f.write("=" * 60 + "\n")
        f.write("数据迁移报告\n")
        f.write("=" * 60 + "\n\n")
        f.write(f"生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
        
        # 统计剧集数据
        shows_file = os.path.join(OUTPUT_DIR, "shows.csv")
        if os.path.exists(shows_file):
            df = pd.read_csv(shows_file)
            f.write(f"剧集数据统计:\n")
            f.write(f"  总记录数: {len(df)}\n")
            f.write(f"  列数: {len(df.columns)}\n")
            f.write(f"  列名: {', '.join(df.columns.tolist())}\n\n")
        
        # 统计剧集详情
        episodes_file = os.path.join(OUTPUT_DIR, "episodes.csv")
        if os.path.exists(episodes_file):
            df = pd.read_csv(episodes_file)
            f.write(f"剧集详情统计:\n")
            f.write(f"  总记录数: {len(df)}\n")
            f.write(f"  列数: {len(df.columns)}\n")
            f.write(f"  列名: {', '.join(df.columns.tolist())}\n\n")
        
        f.write("=" * 60 + "\n")
        f.write("下一步:\n")
        f.write("1. 检查导出的CSV文件\n")
        f.write("2. 运行Go导入脚本: go run scripts/migrate/import.go\n")
        f.write("3. 验证数据完整性\n")
        f.write("=" * 60 + "\n")
    
    print(f"\n✅ 迁移报告已生成: {report_file}")


def main():
    """主函数"""
    print("=" * 60)
    print("Excel to CSV 迁移脚本")
    print("=" * 60)
    print()
    
    # 导出剧集数据
    export_shows_to_csv()
    
    # 导出剧集详情
    export_episodes_to_csv()
    
    # 生成报告
    generate_migration_report()
    
    print("\n" + "=" * 60)
    print("数据导出完成!")
    print("=" * 60)
    print(f"\n输出目录: {os.path.abspath(OUTPUT_DIR)}")
    print("\n请检查导出的CSV文件,然后运行Go导入脚本")


if __name__ == "__main__":
    main()
