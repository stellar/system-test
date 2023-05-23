This is a placeholder file only.

the system-teset/js-soroban-client directory can be used if you want to build system-test 
with a local file path for the js-soroban-client project rather than pulling it from remote 
npm or gh ref.

js-soroban-client$ `yarn build`

then copy the entire js-soroban-client directory over the top of system-test/js-soroban-client, 
or soft link the directories, but be careful in soft link case as you don't want to accidentally delete, etc.

once you have the system-teset/js-soroban-client ready, then build system-test and trigger it to use with:

```
make .... JS_SOROBAN_CLIENT_NPM_VERSION=file:/home/tester/js-soroban-client build

```

system-teset/js-soroban-client is copied to the `/home/tester/js-soroban-client` path in the docker image.
