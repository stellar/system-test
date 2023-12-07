**This is a placeholder file only.**

The `system-test/js-stellar-sdk` directory can be used if you want to build `system-test` with a local file path for the [stellar/js-stellar-sdk](https://github.com/stellar/js-stellar-sdk) project rather than pulling it from remote
npm or GitHub ref. 

This can also be used as alternative when trying to use a GitHub ref url but getting the dreaded `file appears to be corrupt: ENOENT: no such file or directory` during build.


First, clone/check-out a version of js-stellar-sdk from GH locally, it should be clean, remove any node-modules/yarn.lock.


Then, copy the entire `js-stellar-sdk` directory over the top of `system-test/js-stellar-sdk`, this directory will be copied to the `/home/tester/js-stellar-sdk` path in the docker image.

Then build `system-test` and trigger it to compile js sdk using the local files with:

```bash
make .... JS_STELLAR_SDK_NPM_VERSION=file:/home/tester/js-stellar-sdk build
```


