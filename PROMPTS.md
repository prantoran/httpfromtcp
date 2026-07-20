
## Create WASM binary

I want to host the Golang server in cmd/httpserver/main.go as a WebAssembly binary in a way that I can upload the compiled WASM binaries to my static portfolio website (https://prantoran.github.io) and serve the binary to any browser, which will start the Golang server inside the browser and the user can query the Golang server by sending HTTP requests to the Golang server in the browser and using Curl from the terminal. Currently the Golang server supports HTTP/1.1, so browser and terminal curl requests should be HTTP/1.1. 

Add the WebAssembly build system using Makefile and CMake. Use Emscripten as the compiler toolchain. Add scripts to build, compile, run, and test the WebAssembly binary locally. Save the output WASM binary to the `wasm` directory in the root directory.

Add the WebAssembly build/compile and run/local testing instructions and scripts/commands in /docs/WASM.md.

Add descriptive comments in the code to explain the WebAssembly build process and usage, and why the steps and the changes in `main.go` are needed for WebAssembly compatibility. Also add a section in `main.go` about the current limitations for WebAssembly compatibility, and what needs to be done to improve WebAssembly compatibility.

<<Generates implementation plan v1>>

Can you update the implementation plan to add an index.html that will instantiate the Golang server as WASM binary in the background and provide an UI to simulate/make CURL requests to the Golang server, i.e. Like the feel of running cmd command in a terminal. Also add instructions on how to make the curl requests using the UI, i.e. sample curl requests to the Golang server using localcal (not sure).

