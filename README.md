# ast-playground
Ast travel for fun

## Go AST

```
# use go ast to detect cycle import conflicts.
$ cd golang
$ go build
$ ./golang
[CYCLE IMPORT] function call found on pkg B line 20: A.DoSthReplyOnB
refered pkg A usage: [B.Flag]
Boom!
```
