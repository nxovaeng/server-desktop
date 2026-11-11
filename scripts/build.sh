#!/usr/bin/env bash

#
# Xray Panel Desktop (Wails v2) 本地 / CI 自动构建与打包脚本
# 用法:
#   ./scripts/build.sh [--version v0.1.0] [--platform linux/amd64] [--out dist] [--package] [--skip-test]
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

VERSION=""
PLATFORM=""
OUT_DIR="dist"
DO_PACKAGE="false"
SKIP_TEST="false"
EXTRA_TAGS=""
PKG_PREFIX="server-desktop"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --out)
            OUT_DIR="$2"
            shift 2
            ;;
        --tags)
            EXTRA_TAGS="$2"
            shift 2
            ;;
        --package)
            DO_PACKAGE="true"
            shift
            ;;
        --skip-test)
            SKIP_TEST="true"
            shift
            ;;
        -h|--help)
            sed -n '4,7p' "$0"
            exit 0
            ;;
        *)
            echo "未知参数: $1" >&2
            exit 1
            ;;
    esac
done

cd "${REPO_DIR}"

GOPATH_BIN="$(go env GOPATH 2>/dev/null || echo "${HOME}/go")/bin"
if [[ -d "${GOPATH_BIN}" && ":${PATH}:" != *":${GOPATH_BIN}:"* ]]; then
    export PATH="${GOPATH_BIN}:${PATH}"
fi

if [[ -z "${VERSION}" ]]; then
    VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo 'v0.1.0')"
fi

OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_NAME="$(uname -m)"
case "${ARCH_NAME}" in
    x86_64|amd64) ARCH_NORM="amd64" ;;
    aarch64|arm64) ARCH_NORM="arm64" ;;
    *) ARCH_NORM="${ARCH_NAME}" ;;
esac

if [[ -z "${PLATFORM}" ]]; then
    PLATFORM="${OS_NAME}/${ARCH_NORM}"
fi

BUILD_TAGS="${EXTRA_TAGS}"
if [[ "${OS_NAME}" == "linux" && -z "${BUILD_TAGS}" ]]; then
    if command -v pkg-config &>/dev/null && pkg-config --exists webkit2gtk-4.1; then
        BUILD_TAGS="webkit2_41"
        echo "==> 检测到 webkit2gtk-4.1，自动启用 Go build tag: webkit2_41"
    fi
fi

GO_TAG_ARGS=()
WAILS_TAG_ARGS=()
if [[ -n "${BUILD_TAGS}" ]]; then
    GO_TAG_ARGS=(-tags "${BUILD_TAGS}")
    WAILS_TAG_ARGS=(-tags "${BUILD_TAGS}")
fi

if [[ ! -d "frontend/node_modules" ]]; then
    echo "==> 安装前端依赖 (npm ci)"
    (cd frontend && npm ci)
fi

if [[ "${SKIP_TEST}" != "true" ]]; then
    echo "==> 运行前端类型检查与 Go 单元测试"
    (cd frontend && npm run build)
    go vet "${GO_TAG_ARGS[@]}" ./...
    go test "${GO_TAG_ARGS[@]}" -count=1 ./...
fi

if ! command -v wails &>/dev/null; then
    echo "错误: 未找到 wails CLI，请先执行: go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
    exit 1
fi

echo "==> 执行 wails build (platform=${PLATFORM}, version=${VERSION})"
wails build -clean -trimpath -platform "${PLATFORM}" "${WAILS_TAG_ARGS[@]}"

if [[ "${DO_PACKAGE}" == "true" ]]; then
    mkdir -p "${OUT_DIR}"
    BIN_FILE=""
    if [[ -f "build/bin/xray-panel-desktop" ]]; then
        BIN_FILE="build/bin/xray-panel-desktop"
    else
        BIN_FILE="$(find build/bin -maxdepth 1 -type f -executable | head -n1 || true)"
    fi

    if [[ -z "${BIN_FILE}" ]]; then
        echo "错误: 未在 build/bin/ 下找到可打包的构建产物" >&2
        exit 1
    fi

    PLAT_SLUG="${PLATFORM//\//-}"
    PKG_NAME="${PKG_PREFIX}-${VERSION}-${PLAT_SLUG}"
    PKG_DIR="${OUT_DIR}/${PKG_NAME}"
    rm -rf "${PKG_DIR}"
    mkdir -p "${PKG_DIR}"

    cp "${BIN_FILE}" "${PKG_DIR}/"
    chmod +x "${PKG_DIR}/$(basename "${BIN_FILE}")"
    [[ -f README.md ]] && cp README.md "${PKG_DIR}/"

    (
        cd "${OUT_DIR}"
        tar czf "${PKG_NAME}.tar.gz" "${PKG_NAME}"
        if command -v sha256sum &>/dev/null; then
            sha256sum "${PKG_NAME}.tar.gz" > "${PKG_NAME}.tar.gz.sha256"
        else
            shasum -a 256 "${PKG_NAME}.tar.gz" > "${PKG_NAME}.tar.gz.sha256"
        fi
        rm -rf "${PKG_NAME}"
    )
    echo "==> 归档完成: ${OUT_DIR}/${PKG_NAME}.tar.gz"
    ls -lh "${OUT_DIR}/${PKG_NAME}.tar.gz"*
fi
