#!/bin/bash

set -e
echo
echo '开始测试...'

echo "zclog 测试用例"
cd zclog || exit
go test
cd ../

echo
read -rp "zclog 测试用例 结束，按下任意按键继续..." -n 1
echo

cd benchtest || exit
go test -bench=.
cd ../

echo
read -rp "benchtest 测试用例 结束，按下任意按键继续..." -n 1
echo

echo '测试结束...'
echo