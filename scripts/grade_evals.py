"""
due-skills 评估脚本 - 自动评分测试用例
"""

import json
import os
import re

WORKSPACE_DIR = r"E:\Project\hzInfinity\skills\due\due-skills-workspace\iteration-1"

# 评估配置
EVALS = [
    "websocket-gate-server",
    "node-router-handler",
    "mesh-microservice-setup",
    "chat-room-complete",
    "redis-eventbus-config"
]

def read_file(path):
    """读取文件内容"""
    try:
        with open(path, 'r', encoding='utf-8') as f:
            return f.read()
    except Exception as e:
        return None

def check_contains(code, check_str):
    """检查代码是否包含指定字符串"""
    if code is None:
        return False
    return check_str in code

def check_count(code, check_str, min_count):
    """检查代码中指定字符串出现的次数"""
    if code is None:
        return False
    count = code.count(check_str)
    return count >= min_count

def grade_assertions(code, assertions):
    """对单个文件的所有断言进行评分"""
    results = []
    passed_count = 0

    for assertion in assertions:
        assertion_type = assertion.get("type", "code_contains")
        check_str = assertion.get("check", "")
        passed = False

        if assertion_type == "code_contains":
            passed = check_contains(code, check_str)
        elif assertion_type == "code_count":
            min_count = assertion.get("min_count", 1)
            passed = check_count(code, check_str, min_count)

        results.append({
            "name": assertion["name"],
            "text": assertion["description"],
            "passed": passed,
            "evidence": f"检查：{check_str}" if passed else f"未找到：{check_str}"
        })

        if passed:
            passed_count += 1

    return results, passed_count, len(assertions)

def grade_eval(eval_name, assertions):
    """评分单个评估用例"""
    eval_dir = os.path.join(WORKSPACE_DIR, eval_name)

    # 查找输出文件
    with_skill_files = []
    without_skill_files = []

    # with_skill 目录
    with_skill_output = os.path.join(eval_dir, "with_skill", "outputs")
    if os.path.exists(with_skill_output):
        for root, dirs, files in os.walk(with_skill_output):
            for file in files:
                if file.endswith(('.go', '.py', '.js', '.ts')):
                    with_skill_files.append(os.path.join(root, file))

    # without_skill 目录
    without_skill_output = os.path.join(eval_dir, "without_skill", "outputs")
    if os.path.exists(without_skill_output):
        for root, dirs, files in os.walk(without_skill_output):
            for file in files:
                if file.endswith(('.go', '.py', '.js', '.ts')):
                    without_skill_files.append(os.path.join(root, file))

    # 对于 chat-room-complete，需要合并多个文件
    def merge_go_files(file_list):
        content = ""
        for f in file_list:
            file_content = read_file(f)
            if file_content:
                content += f"\n// File: {f}\n" + file_content
        return content

    with_skill_code = merge_go_files(with_skill_files)
    without_skill_code = merge_go_files(without_skill_files)

    # 评分
    with_results, with_passed, with_total = grade_assertions(with_skill_code, assertions)
    without_results, without_passed, without_total = grade_assertions(without_skill_code, assertions)

    return {
        "with_skill": {
            "passed": with_passed,
            "total": with_total,
            "pass_rate": with_passed / with_total if with_total > 0 else 0,
            "results": with_results
        },
        "without_skill": {
            "passed": without_passed,
            "total": without_total,
            "pass_rate": without_passed / without_total if without_total > 0 else 0,
            "results": without_results
        }
    }

def main():
    # 加载 evals.json
    evals_path = r"E:\Project\hzInfinity\skills\due\evals\evals.json"
    with open(evals_path, 'r', encoding='utf-8') as f:
        evals_data = json.load(f)

    all_results = {}
    total_with_passed = 0
    total_with_all = 0
    total_without_passed = 0
    total_without_all = 0

    for eval_item in evals_data["evals"]:
        eval_name = eval_item["name"]
        assertions = eval_item.get("assertions", [])

        print(f"\n正在评估：{eval_name}")
        print(f"  断言数量：{len(assertions)}")

        result = grade_eval(eval_name, assertions)
        all_results[eval_name] = result

        total_with_passed += result["with_skill"]["passed"]
        total_with_all += result["with_skill"]["total"]
        total_without_passed += result["without_skill"]["passed"]
        total_without_all += result["without_skill"]["total"]

        print(f"  With Skill:    {result['with_skill']['passed']}/{result['with_skill']['total']} ({result['with_skill']['pass_rate']*100:.1f}%)")
        print(f"  Without Skill: {result['without_skill']['passed']}/{result['without_skill']['total']} ({result['without_skill']['pass_rate']*100:.1f}%)")

    # 输出汇总
    print("\n" + "="*60)
    print("评估汇总")
    print("="*60)
    print(f"With Skill:    {total_with_passed}/{total_with_all} ({total_with_passed/total_with_all*100:.1f}%)")
    print(f"Without Skill: {total_without_passed}/{total_without_all} ({total_without_passed/total_without_all*100:.1f}%)")

    # 保存 grading.json 到每个 eval 目录
    for eval_item in evals_data["evals"]:
        eval_name = eval_item["name"]
        eval_dir = os.path.join(WORKSPACE_DIR, eval_name)

        # with_skill grading
        with_grading_path = os.path.join(eval_dir, "with_skill", "grading.json")
        with_grading_data = {
            "eval_id": eval_item["id"],
            "eval_name": eval_name,
            "assertions": all_results[eval_name]["with_skill"]["results"]
        }
        os.makedirs(os.path.dirname(with_grading_path), exist_ok=True)
        with open(with_grading_path, 'w', encoding='utf-8') as f:
            json.dump(with_grading_data, f, indent=2, ensure_ascii=False)

        # without_skill grading
        without_grading_path = os.path.join(eval_dir, "without_skill", "grading.json")
        without_grading_data = {
            "eval_id": eval_item["id"],
            "eval_name": eval_name,
            "assertions": all_results[eval_name]["without_skill"]["results"]
        }
        os.makedirs(os.path.dirname(without_grading_path), exist_ok=True)
        with open(without_grading_path, 'w', encoding='utf-8') as f:
            json.dump(without_grading_data, f, indent=2, ensure_ascii=False)

    # 生成 benchmark.json
    benchmark_data = {
        "skill_name": "due-skills",
        "iteration": 1,
        "with_skill": {
            "pass_rate": {
                "mean": total_with_passed / total_with_all * 100 if total_with_all > 0 else 0,
                "stddev": 0
            },
            "time": {
                "mean": 0,
                "stddev": 0
            },
            "tokens": {
                "mean": 0,
                "stddev": 0
            }
        },
        "without_skill": {
            "pass_rate": {
                "mean": total_without_passed / total_without_all * 100 if total_without_all > 0 else 0,
                "stddev": 0
            },
            "time": {
                "mean": 0,
                "stddev": 0
            },
            "tokens": {
                "mean": 0,
                "stddev": 0
            }
        },
        "delta": {
            "pass_rate": (total_with_passed / total_with_all - total_without_passed / total_without_all) * 100 if total_with_all > 0 and total_without_all > 0 else 0,
            "time": 0,
            "tokens": 0
        },
        "evals": []
    }

    for eval_name, result in all_results.items():
        benchmark_data["evals"].append({
            "name": eval_name,
            "with_skill": {
                "pass_rate": result["with_skill"]["pass_rate"] * 100,
                "passed": result["with_skill"]["passed"],
                "total": result["with_skill"]["total"]
            },
            "without_skill": {
                "pass_rate": result["without_skill"]["pass_rate"] * 100,
                "passed": result["without_skill"]["passed"],
                "total": result["without_skill"]["total"]
            }
        })

    benchmark_path = os.path.join(WORKSPACE_DIR, "benchmark.json")
    with open(benchmark_path, 'w', encoding='utf-8') as f:
        json.dump(benchmark_data, f, indent=2, ensure_ascii=False)

    print(f"\nBenchmark 数据已保存到：{benchmark_path}")

    # 生成 benchmark.md
    md_content = f"""# due-skills 评估报告 (Iteration 1)

## 汇总结果

| 配置 | 通过率 | 通过数 | 总断言数 |
|------|--------|--------|----------|
| With Skill | {total_with_passed/total_with_all*100:.1f}% | {total_with_passed} | {total_with_all} |
| Without Skill | {total_without_passed/total_without_all*100:.1f}% | {total_without_passed} | {total_without_all} |
| **提升** | **{(total_with_passed/total_with_all - total_without_passed/total_without_all)*100:+.1f}%** | - | - |

## 各评估用例详情

"""

    for eval_name, result in all_results.items():
        with_rate = result["with_skill"]["pass_rate"] * 100
        without_rate = result["without_skill"]["pass_rate"] * 100
        delta = with_rate - without_rate

        md_content += f"""### {eval_name}

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | {with_rate:.1f}% | {result['with_skill']['passed']}/{result['with_skill']['total']} |
| Without Skill | {without_rate:.1f}% | {result['without_skill']['passed']}/{result['without_skill']['total']} |
| **提升** | **{delta:+.1f}%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
"""
        for i, assertion in enumerate(result["with_skill"]["results"]):
            with_status = "✓" if assertion["passed"] else "✗"
            without_status = "✓" if result["without_skill"]["results"][i]["passed"] else "✗"
            md_content += f"| {assertion['name']} | {with_status} | {without_status} |\n"

        md_content += "\n"

    md_path = os.path.join(WORKSPACE_DIR, "benchmark.md")
    with open(md_path, 'w', encoding='utf-8') as f:
        f.write(md_content)

    print(f"评估报告已保存到：{md_path}")

if __name__ == "__main__":
    main()
