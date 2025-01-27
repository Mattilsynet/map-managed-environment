## requirements

tinygo v0.33  
wit-deps(optional) : https://github.com/bytecodealliance/wit-deps  
wit-bindgen-go: https://github.com/bytecodealliance/wasm-tools-go/tree/main (clone and go install from cmd/)  

## to make it work
1. go get github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go@v0.3.1
1. go generate ./... 
2. go mod tidy
3. wash build
4. wash up
5. wash app deploy wadm.yaml
