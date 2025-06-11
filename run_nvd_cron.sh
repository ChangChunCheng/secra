#!/bin/bash
set -e

# \u5207\u63db Go \u7248\u672c\uff08\u8996\u958b\u767c\u74b0\u5883\u9700\u8981\uff09
source ~/.gvm/scripts/gvm
gvm use go1.23.8

# \u8a2d\u5b9a\u74b0\u5883\u8b8a\u6578\uff08\u5982\u679c\u4e0d\u662f\u9760 .env\uff09
export YEAR=$(date +%Y)
export SOURCE=nvd-cve

# \u53ef\u6307\u5b9a Make target\uff0c\u9810\u8a2d\u70ba import-nvd-v1
MAKE_TARGET="${MAKE_TARGET:-import-nvd-v1}"

# \u5207\u5230\u5c08\u6848\u6839\u76ee\u9304
cd "$(dirname "$0")/.."

# \u57f7\u884c make \u4efb\u52d9
make "$MAKE_TARGET"
