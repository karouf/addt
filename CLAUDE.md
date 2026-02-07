- Golang code files should typically be no longer then 300 lines, max 400
- When you suggest code, also update or add the tests
- When you change code also update the documentation
- This tool is meant for developers - making it easy to use
- In the README I want to have the quick path first and the more advanced further
- When I change cli commands , make sure the completion and the docs are up to date
- When a config setting has multiple values , consider splitting them into different code files

- podman tmpfs has no userid and guid params
- podman can't access sockets 
 On macOS, podman runs containers inside a Linux VM. Unix sockets can't be shared
   from macOS into the VM via virtiofs, so mounting the SSH proxy socket fails with statfs 
  ... operation not supported.       