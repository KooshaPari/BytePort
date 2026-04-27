package lib


/* A Basic NVMS IAC Config to run say FixIt (Svelte/Gin) we'd need 2 running services for our system to function, actually, ideally 3. Our frontend hosted on an open public port, our backend similarly, and our postgres privately. We want to host all of this on a single microvm instance wherever compute is needed, 
and moreover this configuration needs to be such that any other program with the same file structure and build commands could theoretically be deployed on aws via this file, as such this config also needs to directly map to the aws services we need so that each application is fully configured on deploy. 

Fixit is a todolist app built on svelte, gin, and a sqlite DB (postgres in byteport)



*/