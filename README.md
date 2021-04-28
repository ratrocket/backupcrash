# backupcrash

**UPDATE**: The way around all this is to use sqlite's built in `VACUUM
INTO` feature.  See my own sqlutil library for the 2 line backup
implementation...

Illustrates a crash I'm getting using the backup features of
[crawshaw.io/sqlite](https://github.com/crawshaw/sqlite) (and
crawshaw.io/sqlitex).

This is supposed to be a "minimal example" (I hope, haha!).

I ran into the situation described here in a command line tool that
backs up an sqlite database before making some changes to its structure.

## Environment

```
$ go version
go version go1.16.3 linux/amd64

$ gcc --version
gcc (Ubuntu 10.3.0-1ubuntu1) 10.3.0

$ uname -a
Linux crash0 5.11.0-16-generic #17-Ubuntu SMP Wed Apr 14 20:12:43 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux

$ cat /etc/issue
Ubuntu 21.04
```

## Building, running, cleaning up

Clean & build: `make` (does `make clean` and `go build`)

Run: `./backupcrash` (creates database files in project root)

Clean up database files "manually" if you need to with `make clean`

To exercise the "way #2" backup method (see next section for what this
even means...), edit the file and comment out the 4 lines that comprise
"way #1" (labeled "WAY ONE").

## The setup

The program creates a database ("database.db") with a very minimal
"products" table, inserts two phony products, then attempts to back up
the database in two different ways to two distinct files ("backup1.db",
"backup2.db").  After the backups, it would insert another product and
check that the primary database has 3 products and the first backup has
2 products.  "After the backups" is never attained, however -- that's
the problem :).

Way #1 uses
[BackupToDB](https://pkg.go.dev/crawshaw.io/sqlite#Conn.BackupToDB)
which looks to me like the easy, one-stop-shop method of back up.

Way #2 uses the building blocks in package sqlite that BackupToDB itself
uses: `BackupInit`, `Step`, `Finish`.  ([Source of BackupToDB for
reference](https://github.com/crawshaw/sqlite/blob/v0.3.2/backup.go#L46).)

## Questions

1. Am I using these functions wrong?  (Could well be!)

1. Is there a more up-to-date or more recommended way to do backups with
   package sqlite?

## Results

### Summary

Way #1's headline is "free(): invalid pointer"

Way #2's headline is "free(): invalid size"

In both cases, the error, as far as package sqlite is concerned, comes
from [backup.go, line
77](https://github.com/crawshaw/sqlite/blob/v0.3.2/backup.go#L77), in
function BackupInit when `b.ptr == nil` evaluates to true, and the
deferred function `setCDB` is called, from [backup.go, line
93](https://github.com/crawshaw/sqlite/blob/v0.3.2/backup.go#L93).

**I suspect that `b.ptr` being nil is the root of the problem.**  The
rest (the return on line 77 which kicks off the deferred function call
that triggers the panic) is just the aftermath.

So the problem is really [backup.go, line
74](https://github.com/crawshaw/sqlite/blob/v0.3.2/backup.go#L74) where
the call `C.sqlite3_backup_init(dst.conn, dstCDB, src.conn, srcCDB)`
returns nil (and shouldn't).

#### Note on mac OS

I was able to run the program on a mac with an outdated go version.  The
results seem basically the same.  The first line of output for _both_ ways is

```
backupcrash(99215,0xc38e5c0) malloc: *** error for object 0xc0000a60a0:
pointer being freed was not allocated
```

(This is different from on linux, where each way has a _slightly_
different message...)

The go version used on the mac is (very outdated, I know...)

```
$ go version
go version go1.13.12 darwin/amd64
```

(END note on mac OS)

What follows is just the crash output (on linux, the system detailed at
the top of this document).

### Way #1, gory details

```
     1	free(): invalid pointer
     2	SIGABRT: abort
     3	PC=0x7f4c3024bfbb m=0 sigcode=18446744073709551610
     4
     5	goroutine 0 [idle]:
     6	runtime: unknown pc 0x7f4c3024bfbb
     7	stack: frame={sp:0x7ffd14d482c0, fp:0x0} stack=[0x7ffd14549718,0x7ffd14d48750)
     8	00007ffd14d481c0:  00007ffd14d481d0  3717c3175031a100
     9	00007ffd14d481d0:  00000000223ff73b  00000000004d0fb6
    10	00007ffd14d481e0:  00000000223ff73b  00000000022c5a76
    11	00007ffd14d481f0:  0000000000000000  0000000000000000
    12	00007ffd14d48200:  00000000022c59f0  00000000004eaf43
    13	00007ffd14d48210:  0000000000000806  0000000000000000
    14	00007ffd14d48220:  00000000022c5700  0000000000000200
    15	00007ffd14d48230:  0000000000000000  0000000000000000
    16	00007ffd14d48240:  00000000022c5ad8  3717c3175031a100
    17	00007ffd14d48250:  00007ffd14d48330  00007ffd14d48290
    18	00007ffd14d48260:  00007ffd14d484c4  0000000000000000
    19	00007ffd14d48270:  0000000000000000  0000000000000200
    20	00007ffd14d48280:  00000000022c5a54  00000000004ed7ad
    21	00007ffd14d48290:  64732f656d6f682f  70756b6361622f66
    22	00007ffd14d482a0:  6162006873617263  62642e3170756b63
    23	00007ffd14d482b0:  6c616e72756f6a2d  0000000000530900
    24	00007ffd14d482c0: <0000000000000000  00000000022c23c8
    25	00007ffd14d482d0:  0000000000000002  3717c3175031a100
    26	00007ffd14d482e0:  00000000022c5768  ffffffffffffffff
    27	00007ffd14d482f0:  00000000022c5028  00000000004e42b7
    28	00007ffd14d48300:  00000000005fd1a9  ffffffffffffffff
    29	00007ffd14d48310:  00007ffd14d48330  00000000004ef72f
    30	00007ffd14d48320:  00000000022eaac0  3717c3175031a100
    31	00007ffd14d48330:  00000000022c5028  00000000022d4de8
    32	00007ffd14d48340:  fffffffe7fffffff  ffffffffffffffff
    33	00007ffd14d48350:  ffffffffffffffff  ffffffffffffffff
    34	00007ffd14d48360:  ffffffffffffffff  ffffffffffffffff
    35	00007ffd14d48370:  ffffffffffffffff  ffffffffffffffff
    36	00007ffd14d48380:  ffffffffffffffff  ffffffffffffffff
    37	00007ffd14d48390:  ffffffffffffffff  ffffffffffffffff
    38	00007ffd14d483a0:  ffffffffffffffff  ffffffffffffffff
    39	00007ffd14d483b0:  ffffffffffffffff  ffffffffffffffff
    40	runtime: unknown pc 0x7f4c3024bfbb
    41	stack: frame={sp:0x7ffd14d482c0, fp:0x0} stack=[0x7ffd14549718,0x7ffd14d48750)
    42	00007ffd14d481c0:  00007ffd14d481d0  3717c3175031a100
    43	00007ffd14d481d0:  00000000223ff73b  00000000004d0fb6
    44	00007ffd14d481e0:  00000000223ff73b  00000000022c5a76
    45	00007ffd14d481f0:  0000000000000000  0000000000000000
    46	00007ffd14d48200:  00000000022c59f0  00000000004eaf43
    47	00007ffd14d48210:  0000000000000806  0000000000000000
    48	00007ffd14d48220:  00000000022c5700  0000000000000200
    49	00007ffd14d48230:  0000000000000000  0000000000000000
    50	00007ffd14d48240:  00000000022c5ad8  3717c3175031a100
    51	00007ffd14d48250:  00007ffd14d48330  00007ffd14d48290
    52	00007ffd14d48260:  00007ffd14d484c4  0000000000000000
    53	00007ffd14d48270:  0000000000000000  0000000000000200
    54	00007ffd14d48280:  00000000022c5a54  00000000004ed7ad
    55	00007ffd14d48290:  64732f656d6f682f  70756b6361622f66
    56	00007ffd14d482a0:  6162006873617263  62642e3170756b63
    57	00007ffd14d482b0:  6c616e72756f6a2d  0000000000530900
    58	00007ffd14d482c0: <0000000000000000  00000000022c23c8
    59	00007ffd14d482d0:  0000000000000002  3717c3175031a100
    60	00007ffd14d482e0:  00000000022c5768  ffffffffffffffff
    61	00007ffd14d482f0:  00000000022c5028  00000000004e42b7
    62	00007ffd14d48300:  00000000005fd1a9  ffffffffffffffff
    63	00007ffd14d48310:  00007ffd14d48330  00000000004ef72f
    64	00007ffd14d48320:  00000000022eaac0  3717c3175031a100
    65	00007ffd14d48330:  00000000022c5028  00000000022d4de8
    66	00007ffd14d48340:  fffffffe7fffffff  ffffffffffffffff
    67	00007ffd14d48350:  ffffffffffffffff  ffffffffffffffff
    68	00007ffd14d48360:  ffffffffffffffff  ffffffffffffffff
    69	00007ffd14d48370:  ffffffffffffffff  ffffffffffffffff
    70	00007ffd14d48380:  ffffffffffffffff  ffffffffffffffff
    71	00007ffd14d48390:  ffffffffffffffff  ffffffffffffffff
    72	00007ffd14d483a0:  ffffffffffffffff  ffffffffffffffff
    73	00007ffd14d483b0:  ffffffffffffffff  ffffffffffffffff
    74
    75	goroutine 1 [syscall]:
    76	runtime.cgocall(0x4c96a0, 0xc000043ce0, 0x5f99b8)
    77		/usr/local/go/src/runtime/cgocall.go:154 +0x5b fp=0xc000043cb0 sp=0xc000043c78 pc=0x40cb1b
    78	crawshaw.io/sqlite._Cfunc_free(0xc0000100b8)
    79		_cgo_gotypes.go:428 +0x3c fp=0xc000043ce0 sp=0xc000043cb0 pc=0x4b67fc
    80	crawshaw.io/sqlite.setCDB.func2.1(0xc0000100b8)
    81		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:93 +0x4d fp=0xc000043d10 sp=0xc000043ce0 pc=0x4beaad
    82	crawshaw.io/sqlite.setCDB.func2()
    83		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:93 +0x2a fp=0xc000043d28 sp=0xc000043d10 pc=0x4beaea
    84	crawshaw.io/sqlite.(*Conn).BackupInit(0xc000088000, 0x5d619e, 0xb, 0x0, 0x0, 0xc000088460, 0x0, 0x5f99b8, 0xc0000520c0)
    85		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:77 +0x230 fp=0xc000043da8 sp=0xc000043d28 pc=0x4b8cb0
    86	crawshaw.io/sqlite.(*Conn).BackupToDB(0xc000088000, 0x5d619e, 0xb, 0x5d5e85, 0xa, 0xc000088460, 0x0, 0x0)
    87		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:55 +0x114 fp=0xc000043e30 sp=0xc000043da8 pc=0x4b8974
    88	main.main()
    89		/home/sdf/backupcrash/main.go:43 +0x326 fp=0xc000043f88 sp=0xc000043e30 pc=0x4c85e6
    90	runtime.main()
    91		/usr/local/go/src/runtime/proc.go:225 +0x256 fp=0xc000043fe0 sp=0xc000043f88 pc=0x440716
    92	runtime.goexit()
    93		/usr/local/go/src/runtime/asm_amd64.s:1371 +0x1 fp=0xc000043fe8 sp=0xc000043fe0 pc=0x4720a1
    94
    95	goroutine 6 [select]:
    96	crawshaw.io/sqlite.(*Conn).SetInterrupt.func1(0xc000068120, 0xc000088000, 0xc000068180)
    97		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:313 +0x70
    98	created by crawshaw.io/sqlite.(*Conn).SetInterrupt
    99		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:312 +0x17a
   100
   101	rax    0x0
   102	rbx    0x7f4c30209c00
   103	rcx    0x7f4c3024bfbb
   104	rdx    0x0
   105	rdi    0x2
   106	rsi    0x7ffd14d482c0
   107	rbp    0x7ffd14d48610
   108	rsp    0x7ffd14d482c0
   109	r8     0x0
   110	r9     0x7ffd14d482c0
   111	r10    0x8
   112	r11    0x246
   113	r12    0x7ffd14d48530
   114	r13    0x10
   115	r14    0x7f4c30575000
   116	r15    0x1
   117	rip    0x7f4c3024bfbb
   118	rflags 0x246
   119	cs     0x33
   120	fs     0x0
   121	gs     0x0
```

### Way #2, gory details

(results had by commenting out "way one" in the code, `make clean && go
build`)

```
     1	free(): invalid size
     2	SIGABRT: abort
     3	PC=0x7f109c95afbb m=0 sigcode=18446744073709551610
     4
     5	goroutine 0 [idle]:
     6	runtime: unknown pc 0x7f109c95afbb
     7	stack: frame={sp:0x7ffc9e6a8a50, fp:0x0} stack=[0x7ffc9dea9ea8,0x7ffc9e6a8ee0)
     8	00007ffc9e6a8950:  00007ffc9e6a8960  2b96be81c1980400
     9	00007ffc9e6a8960:  0000000020581d3d  00000000004d0c36
    10	00007ffc9e6a8970:  0000000020581d3d  0000000000d22a76
    11	00007ffc9e6a8980:  0000000000000000  0000000000000000
    12	00007ffc9e6a8990:  0000000000d229f0  00000000004eabc3
    13	00007ffc9e6a89a0:  0000000000000030  0000000000000000
    14	00007ffc9e6a89b0:  0000000000000000  0000000000000000
    15	00007ffc9e6a89c0:  0000000000000000  0000000000000000
    16	00007ffc9e6a89d0:  0000000000000000  00007ffc9e6a8a98
    17	00007ffc9e6a89e0:  00007ffc9e6a8ac0  0000000000650384
    18	00007ffc9e6a89f0:  000000000000016c  000000000000016c
    19	00007ffc9e6a8a00:  00007ffc9e0000c0  000000000045bf25 <runtime.pcvalue+357>
    20	00007ffc9e6a8a10:  0000000000650452  000000000000009e
    21	00007ffc9e6a8a20:  000000000000009e  00007ffc9e6a8a70
    22	00007ffc9e6a8a30:  00007ffc9e6a8a64  0000015000000000
    23	00007ffc9e6a8a40:  0000000000650455  000000000000009b
    24	00007ffc9e6a8a50: <0000000000000000  00007ffc00000001
    25	00007ffc9e6a8a60:  ffffffff00000020  2b96be81c1980400
    26	00007ffc9e6a8a70:  00000000004c8583 <main.main+1347>  ffffffffffffffff
    27	00007ffc9e6a8a80:  0000000000d22028  00000000004e3f37
    28	00007ffc9e6a8a90:  00000000005fc0e9  ffffffffffffffff
    29	00007ffc9e6a8aa0:  00007ffc9e6a8ac0  00000000004ef3af
    30	00007ffc9e6a8ab0:  00000000006ad160  0000000000032f22
    31	00007ffc9e6a8ac0:  0000000000d22028  0000000000d31ce8
    32	00007ffc9e6a8ad0:  fffffffe7fffffff  ffffffffffffffff
    33	00007ffc9e6a8ae0:  ffffffffffffffff  ffffffffffffffff
    34	00007ffc9e6a8af0:  ffffffffffffffff  ffffffffffffffff
    35	00007ffc9e6a8b00:  ffffffffffffffff  ffffffffffffffff
    36	00007ffc9e6a8b10:  ffffffffffffffff  ffffffffffffffff
    37	00007ffc9e6a8b20:  ffffffffffffffff  ffffffffffffffff
    38	00007ffc9e6a8b30:  ffffffffffffffff  ffffffffffffffff
    39	00007ffc9e6a8b40:  ffffffffffffffff  ffffffffffffffff
    40	runtime: unknown pc 0x7f109c95afbb
    41	stack: frame={sp:0x7ffc9e6a8a50, fp:0x0} stack=[0x7ffc9dea9ea8,0x7ffc9e6a8ee0)
    42	00007ffc9e6a8950:  00007ffc9e6a8960  2b96be81c1980400
    43	00007ffc9e6a8960:  0000000020581d3d  00000000004d0c36
    44	00007ffc9e6a8970:  0000000020581d3d  0000000000d22a76
    45	00007ffc9e6a8980:  0000000000000000  0000000000000000
    46	00007ffc9e6a8990:  0000000000d229f0  00000000004eabc3
    47	00007ffc9e6a89a0:  0000000000000030  0000000000000000
    48	00007ffc9e6a89b0:  0000000000000000  0000000000000000
    49	00007ffc9e6a89c0:  0000000000000000  0000000000000000
    50	00007ffc9e6a89d0:  0000000000000000  00007ffc9e6a8a98
    51	00007ffc9e6a89e0:  00007ffc9e6a8ac0  0000000000650384
    52	00007ffc9e6a89f0:  000000000000016c  000000000000016c
    53	00007ffc9e6a8a00:  00007ffc9e0000c0  000000000045bf25 <runtime.pcvalue+357>
    54	00007ffc9e6a8a10:  0000000000650452  000000000000009e
    55	00007ffc9e6a8a20:  000000000000009e  00007ffc9e6a8a70
    56	00007ffc9e6a8a30:  00007ffc9e6a8a64  0000015000000000
    57	00007ffc9e6a8a40:  0000000000650455  000000000000009b
    58	00007ffc9e6a8a50: <0000000000000000  00007ffc00000001
    59	00007ffc9e6a8a60:  ffffffff00000020  2b96be81c1980400
    60	00007ffc9e6a8a70:  00000000004c8583 <main.main+1347>  ffffffffffffffff
    61	00007ffc9e6a8a80:  0000000000d22028  00000000004e3f37
    62	00007ffc9e6a8a90:  00000000005fc0e9  ffffffffffffffff
    63	00007ffc9e6a8aa0:  00007ffc9e6a8ac0  00000000004ef3af
    64	00007ffc9e6a8ab0:  00000000006ad160  0000000000032f22
    65	00007ffc9e6a8ac0:  0000000000d22028  0000000000d31ce8
    66	00007ffc9e6a8ad0:  fffffffe7fffffff  ffffffffffffffff
    67	00007ffc9e6a8ae0:  ffffffffffffffff  ffffffffffffffff
    68	00007ffc9e6a8af0:  ffffffffffffffff  ffffffffffffffff
    69	00007ffc9e6a8b00:  ffffffffffffffff  ffffffffffffffff
    70	00007ffc9e6a8b10:  ffffffffffffffff  ffffffffffffffff
    71	00007ffc9e6a8b20:  ffffffffffffffff  ffffffffffffffff
    72	00007ffc9e6a8b30:  ffffffffffffffff  ffffffffffffffff
    73	00007ffc9e6a8b40:  ffffffffffffffff  ffffffffffffffff
    74
    75	goroutine 1 [syscall]:
    76	runtime.cgocall(0x4c9320, 0xc000043d68, 0x5f8928)
    77		/usr/local/go/src/runtime/cgocall.go:154 +0x5b fp=0xc000043d38 sp=0xc000043d00 pc=0x40cb1b
    78	crawshaw.io/sqlite._Cfunc_free(0xc000010100)
    79		_cgo_gotypes.go:428 +0x3c fp=0xc000043d68 sp=0xc000043d38 pc=0x4b67fc
    80	crawshaw.io/sqlite.setCDB.func2.1(0xc000010100)
    81		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:93 +0x4d fp=0xc000043d98 sp=0xc000043d68 pc=0x4be82d
    82	crawshaw.io/sqlite.setCDB.func2()
    83		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:93 +0x2a fp=0xc000043db0 sp=0xc000043d98 pc=0x4be86a
    84	crawshaw.io/sqlite.(*Conn).BackupInit(0xc000088000, 0x5d5194, 0xb, 0x5d4e85, 0xa, 0xc000088460, 0x0, 0x5f8928, 0xc000052140)
    85		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/backup.go:77 +0x230 fp=0xc000043e30 sp=0xc000043db0 pc=0x4b8a90
    86	main.main()
    87		/home/sdf/backupcrash/main.go:57 +0x41e fp=0xc000043f88 sp=0xc000043e30 pc=0x4c845e
    88	runtime.main()
    89		/usr/local/go/src/runtime/proc.go:225 +0x256 fp=0xc000043fe0 sp=0xc000043f88 pc=0x440716
    90	runtime.goexit()
    91		/usr/local/go/src/runtime/asm_amd64.s:1371 +0x1 fp=0xc000043fe8 sp=0xc000043fe0 pc=0x4720a1
    92
    93	goroutine 6 [select]:
    94	crawshaw.io/sqlite.(*Conn).SetInterrupt.func1(0xc000068120, 0xc000088000, 0xc000068180)
    95		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:313 +0x70
    96	created by crawshaw.io/sqlite.(*Conn).SetInterrupt
    97		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:312 +0x17a
    98
    99	goroutine 7 [runnable]:
   100	crawshaw.io/sqlite.(*Conn).SetInterrupt.func1(0xc000068240, 0xc000088460, 0xc0000682a0)
   101		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:312
   102	created by crawshaw.io/sqlite.(*Conn).SetInterrupt
   103		/home/sdf/go/pkg/mod/crawshaw.io/sqlite@v0.3.2/sqlite.go:312 +0x17a
   104
   105	rax    0x0
   106	rbx    0x7f109c918c00
   107	rcx    0x7f109c95afbb
   108	rdx    0x0
   109	rdi    0x2
   110	rsi    0x7ffc9e6a8a50
   111	rbp    0x7ffc9e6a8da0
   112	rsp    0x7ffc9e6a8a50
   113	r8     0x0
   114	r9     0x7ffc9e6a8a50
   115	r10    0x8
   116	r11    0x246
   117	r12    0x7ffc9e6a8cc0
   118	r13    0x10
   119	r14    0x7f109cc84000
   120	r15    0x1
   121	rip    0x7f109c95afbb
   122	rflags 0x246
   123	cs     0x33
   124	fs     0x0
   125	gs     0x0
```
