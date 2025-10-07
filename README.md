# go socket programming test
使用前需要基於config_example.yaml建立一份config.yaml

## local端測試
*config.yaml的所有IP改為127.0.0.1*

測試所有功能
```
go run test.go all
```
僅輸出client結果
```
go run test.go all -q
```

## 實際部署測試
*根據設備設定好ip後只需執行特定.go就好*