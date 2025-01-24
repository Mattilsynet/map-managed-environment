## requirements

tinygo v0.33  
wit-deps(optional) : https://github.com/bytecodealliance/wit-deps  
wit-bindgen-go: https://github.com/bytecodealliance/wasm-tools-go/tree/main (clone and go install from cmd/)  

## to make it work

1. wash build  
2. wash up (in other terminal)  
3. wash app deploy wadm.yaml  
4. nats req "wasmcloud.echo" "Hello" --server=nats//localhost:4222  
