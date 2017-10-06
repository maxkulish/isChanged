Golang utility to check changes in requirements.txt and package.json

### How to run it
```
./isChanged -npm=/package.json
./isChanged -npm=/../../package.json
```

In package.json checks only two arguments: "dependencies" and "devDependencies".

First check save maps to gob files. Second - compare maps and return Exit.code(10) if changed, Exit.code(11) if equal.

To check pip file requirements.txt
```
./isChanged -pip=/requirements.txt
or
./isChanged -pip=/../../requirements.txt
```

### How to use it
Check if package.json changed remove old node_modules and npm install

Makefile
```
.isChanged -npm=/../js/package.json \
if [ $$? -eq 10 ] ; \
then cd $(JS_DIR) && rm -rf node_modules && npm install ; fi
```
