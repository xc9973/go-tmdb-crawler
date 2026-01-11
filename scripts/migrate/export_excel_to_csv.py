#!/usr/bin/env python3
"""
Excel到CSV导出脚本
将Python版本的Excel数据导出为CSV格式,用于导入Go版本
"""

import pandas as pd
import os
import re
from pathlib import Path

# 配置
PROJECT_ROOT = Path("/Volumes/1disk/爬去")
DATA_DIR = PROJECT_ROOT / "py" / "data"
OUTPUT_DIR = Path(__file__).parent / "output"

# 创建输出目录
OUTPUT_DIR.mkdir(exist_ok=True)

def extract_tmdb_id(filename):
    """从文件名中提取TMDB ID"""
    # 文件名格式: 剧集名称（年份）-123456.xlsx
    match = re.search(r'-(\d+)\.xlsx$', filename)
    if match:
        return int(match.group(1))
    return None

def export_shows_list():
    """导出剧集列表"""
    print("=" * 60)
    print("导出剧集列表...")
    print("=" * 60)
    
    excel_file = DATA_DIR / "剧集列表.xlsx"
    
    if not excel_file.exists():
        print(f"错误: 文件不存在 {excel_file}")
        return False
    
    # 读取Excel
    df = pd.read_excel(excel_file)
    
    # 重命名列以匹配导入程序期望的列名
    column_mapping = {
        'TMDB_ID': 'TMDB_ID',
        '剧集名称': '名称',
        '总季数': '总季数',
        '状态': '状态',
        '最新季度': '最新季度',
        '备注': '备注',
        '下一集更新日期': '下一集更新日期'
    }
    
    # 先重命名列
    df = df.rename(columns=column_mapping)
    
    # 只保留需要的列
    columns_to_keep = ['TMDB_ID', '名称', '状态', '备注']
    df_export = df[columns_to_keep].copy()
    
    # 添加空列以匹配完整结构
    df_export['原名'] = ''
    df_export['简介'] = ''
    df_export['海报路径'] = ''
    df_export['背景路径'] = ''
    df_export['类型'] = ''
    df_export['评分'] = ''
    
    # 重新排列列顺序
    df_export = df_export[[
        'TMDB_ID', '名称', '原名', '状态', '简介', 
        '海报路径', '背景路径', '类型', '评分', '备注'
    ]]
    
    # 保存为CSV
    output_file = OUTPUT_DIR / "shows.csv"
    df_export.to_csv(output_file, index=False, encoding='utf-8-sig')
    
    print(f"✅ 成功导出 {len(df_export)} 条剧集记录")
    print(f"   文件: {output_file}")
    return True

def export_episodes_details():
    """导出剧集详情"""
    print("\n" + "=" * 60)
    print("导出剧集详情...")
    print("=" * 60)
    
    details_dir = DATA_DIR / "剧集详情"
    
    if not details_dir.exists():
        print(f"错误: 目录不存在 {details_dir}")
        return False
    
    # 收集所有剧集详情数据
    all_episodes = []
    
    # 遍历所有Excel文件
    excel_files = list(details_dir.glob("*.xlsx"))
    # 排除.DS_Store
    excel_files = [f for f in excel_files if not f.name.startswith('.')]
    
    print(f"找到 {len(excel_files)} 个剧集详情文件")
    
    for excel_file in excel_files:
        tmdb_id = extract_tmdb_id(excel_file.name)
        if not tmdb_id:
            print(f"  跳过: 无法提取TMDB ID - {excel_file.name}")
            continue
        
        try:
            # 读取Excel
            df = pd.read_excel(excel_file)
            
            # 重命名列
            column_mapping = {
                '季度': '季数',
                '集数': '集数',
                '标题': '名称',
                '简介': '简介',
                '播出日期': '播出日期'
            }
            df = df.rename(columns=column_mapping)
            
            # 添加TMDB ID列
            df['TMDB_ID'] = tmdb_id
            
            # 添加空列
            df['剧照路径'] = ''
            df['评分'] = ''
            
            # 重新排列列顺序
            df = df[[
                'TMDB_ID', '季数', '集数', '名称', '简介', 
                '剧照路径', '播出日期', '评分'
            ]]
            
            all_episodes.append(df)
            
            print(f"  ✅ {excel_file.name}: {len(df)} 集")
            
        except Exception as e:
            print(f"  ❌ 错误: {excel_file.name} - {e}")
            continue
    
    if not all_episodes:
        print("错误: 没有成功导入任何剧集详情")
        return False
    
    # 合并所有数据
    df_all = pd.concat(all_episodes, ignore_index=True)
    
    # 保存为CSV
    output_file = OUTPUT_DIR / "episodes.csv"
    df_all.to_csv(output_file, index=False, encoding='utf-8-sig')
    
    print(f"\n✅ 成功导出 {len(df_all)} 条剧集详情记录")
    print(f"   来自 {len(all_episodes)} 个剧集")
    print(f"   文件: {output_file}")
    return True

def main():
    print("=" * 60)
    print("Excel 到 CSV 导出工具")
    print("=" * 60)
    print()
    
    # 导出剧集列表
    success_shows = export_shows_list()
    
    # 导出剧集详情
    success_episodes = export_episodes_details()
    
    # 总结
    print("\n" + "=" * 60)
    print("导出完成")
    print("=" * 60)
    
    if success_shows and success_episodes:
        print("✅ 所有数据导出成功!")
        print(f"\n输出目录: {OUTPUT_DIR}")
        print("\n下一步:")
        print("1. 启动Go服务: cd go-tmdb-crawler && docker-compose up -d")
        print("2. 运行导入程序: cd scripts/migrate && go run import.go")
    else:
        print("❌ 部分数据导出失败,请检查错误信息")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())
