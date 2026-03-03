#!/bin/bash

read -r json_input

path=$(echo "$json_input" | jq -r '.path // empty')

if [ -z "$path" ]; then
    echo '{"error": "缺少必需参数: path"}'
    exit 1
fi

# SECURITY: 路径验证
# 检查路径遍历
if echo "$path" | grep -q '\.\.'; then
    echo '{"error": "Path traversal detected: .. not allowed"}'
    exit 1
fi

# 设置允许的基础目录
BASE_DIR="${MINDX_WORKSPACE}/documents"
if [ ! -d "$BASE_DIR" ]; then
    BASE_DIR="${MINDX_WORKSPACE}/data"
    if [ ! -d "$BASE_DIR" ]; then
        echo '{"error": "No valid base directory found under MINDX_WORKSPACE (expected documents or data)"}'
        exit 1
    fi
fi

# 如果路径不是绝对路径，则基于BASE_DIR构建完整路径
if [[ "$path" != /* ]]; then
    FULL_PATH="${BASE_DIR}/${path}"
else
    FULL_PATH="$path"
fi

# 规范化路径并验证
normalize_path() {
    local input_path="$1"

    if command -v realpath >/dev/null 2>&1; then
        if REALPATH_OUTPUT=$(realpath -m -- "$input_path" 2>/dev/null); then
            printf '%s\n' "$REALPATH_OUTPUT"
            return 0
        fi
    fi

    if command -v python3 >/dev/null 2>&1; then
        python3 -c "import os,sys; p=sys.argv[1]; p=os.path.join(os.getcwd(),p) if not os.path.isabs(p) else p; print(os.path.normpath(p))" "$input_path"
        return $?
    fi

    return 1
}

REAL_PATH=$(normalize_path "$FULL_PATH") || {
    echo '{"error": "Failed to normalize path"}'
    exit 1
}
REAL_BASE=$(normalize_path "$BASE_DIR") || {
    echo '{"error": "Failed to normalize base directory"}'
    exit 1
}

# 检查规范化后的路径是否仍在基础目录内
if [[ "$REAL_PATH" != "$REAL_BASE" ]] && [[ "$REAL_PATH" != "$REAL_BASE"/* ]]; then
    echo '{"error": "Access denied: path outside allowed directory"}'
    exit 1
fi

# 使用验证后的路径
path="$REAL_PATH"

if [ ! -f "$path" ]; then
    echo "{\"success\": false, \"error\": \"文件不存在: $path\"}"
    exit 1
fi

if [ ! -r "$path" ]; then
    echo "{\"success\": false, \"error\": \"没有文件读取权限: $path\"}"
    exit 1
fi

content=$(cat "$path" 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
    content_escaped=$(printf '%s' "$content" | jq -R -s '.')
    bytes_read=$(wc -c < "$path")
    echo "{\"success\": true, \"path\": \"$path\", \"content\": $content_escaped, \"bytes_read\": $bytes_read}"
else
    error_escaped=$(printf '%s' "$content" | jq -R -s '.')
    echo "{\"success\": false, \"error\": $error_escaped}"
    exit 1
fi
