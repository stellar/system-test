**This is a placeholder file only.**

The `system-test/js-stellar-sdk` directory can be used if you want to build `system-test` with a local file path for the [stellar/js-stellar-sdk](https://github.com/stellar/js-stellar-sdk) project rather than pulling it from remote
npm or GitHub ref. Run,

```bash
js-stellar-sdk$ yarn build
```

Then, copy the entire directory over the top of `system-test/js-stellar-sdk` (or soft link the directories, but be careful in the soft link case as you don't want to accidentally delete, etc.)

Once you have the `system-test/js-stellar-sdk` ready, then build `system-test` and trigger it to use with:

```bash
make .... JS_STELLAR_SDK_NPM_VERSION=file:/home/tester/js-stellar-sdk build
```

Then, `system-test/js-stellar-sdk` is copied to the `/home/tester/js-stellar-sdk` path in the docker image.
