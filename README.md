

# expect
`import "github.com/leemcloughlin/expect"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>
Expect is pure Go (golang) version of the terminal interaction package Expect
common on many Linux systems

A very simple example that calls the rev program to reverse the text in each
line is:


	exp, err := NewExpect("rev")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(5)
	exp.Send("hello\r")
	
	// i will be the index of the argument that matched the input otherwise
	// i will be a negative number showing match fail or error
	i, found, err := exp.Expect("olleh")
	if i == 0 {
		fmt.Println("found", string(found))
	} else {
		fmt.Println("failed ", err)
	}
	exp.Send(EOF)

This package has only been tested on Linux




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [type Expect](#Expect)
  * [func NewExpect(prog string, arg ...string) (*Expect, error)](#NewExpect)
  * [func NewExpectProc(ctx context.Context, prog string, arg ...string) (*Expect, *os.Process, error)](#NewExpectProc)
  * [func Spawn(prog string, arg ...string) (*Expect, error)](#Spawn)
  * [func (exp *Expect) BufStr() string](#Expect.BufStr)
  * [func (exp *Expect) Clear()](#Expect.Clear)
  * [func (exp *Expect) Expect(reOrStrs ...interface{}) (int, []byte, error)](#Expect.Expect)
  * [func (exp *Expect) Expecti(reOrStrs ...interface{}) int](#Expect.Expecti)
  * [func (exp *Expect) Expectp(reOrStrs ...interface{}) int](#Expect.Expectp)
  * [func (exp *Expect) Kill()](#Expect.Kill)
  * [func (exp *Expect) LogUser(on bool)](#Expect.LogUser)
  * [func (exp *Expect) Send(s string) (int, error)](#Expect.Send)
  * [func (exp *Expect) SendL(lines ...string) error](#Expect.SendL)
  * [func (exp *Expect) SendSlow(delay time.Duration, s string) (int, error)](#Expect.SendSlow)
  * [func (exp *Expect) SetCmdOut(cmdOut io.Writer)](#Expect.SetCmdOut)
  * [func (exp *Expect) SetTimeout(timeout time.Duration)](#Expect.SetTimeout)
  * [func (exp *Expect) SetTimeoutSecs(timeout int)](#Expect.SetTimeoutSecs)
* [type ExpectWaitResult](#ExpectWaitResult)

#### <a name="pkg-examples">Examples</a>
* [Expect](#example_Expect)
* [Expect (OwnEcho)](#example_Expect_ownEcho)

#### <a name="pkg-files">Package files</a>
[expect.go](/src/github.com/leemcloughlin/expect/expect.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    // NotFound is returned by Expect for non-matching non-errors (include EOF)
    NotFound = -1

    // TimedOut is returned by Expect on a timeout (along with an error of ETimedOut)
    TimedOut = -2

    // NotStringOrRexgexp is returned if a paramter is not a string or a regexp
    NotStringOrRexgexp = -3
)
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    ETimedOut           = errors.New("TimedOut")
    ENotStringOrRexgexp = errors.New("Not string or regexp")
    EReadError          = errors.New("Read Error")

    // Debug if true will generate vast amounts of internal logging
    Debug = false

    // ExpectInSize is the size of the channel between the expectReader and Expect.
    // If you overflow this than expectReader will block.
    ExpectInSize = 20 * 1024

    // EOF is the terminal EOF character - NOT guaranteed to be correct on all
    // systems nor if the program passed to expect connects to a different
    // platform
    EOF = "\004"
)
```



## <a name="Expect">type</a> [Expect](/src/target/expect.go?s=2105:2883#L85)
``` go
type Expect struct {

    // Expect reads into here. On a successful match all data up the end of the
    // match is deleted shrinking the buffer.
    // See also Clear() and BufStr()
    Buffer *bytes.Buffer

    // On EOF being read from Cmd this is set (and ExpectReader is ended)
    Eof bool

    // Result is filled in asynchronously after the cmd exits
    Result ExpectWaitResult
    // contains filtered or unexported fields
}
```






### <a name="NewExpect">func</a> [NewExpect](/src/target/expect.go?s=3431:3490#L131)
``` go
func NewExpect(prog string, arg ...string) (*Expect, error)
```
NewExpect starts prog, passing any given args, in its own pty.
Note that in order to be non-blocking while reading from the pty this sets
the non-blocking flag and looks for EAGAIN on reads failing.  This has only
been tested on Linux systems.
On prog exiting or being killed Result is filled in shortly after.


### <a name="NewExpectProc">func</a> [NewExpectProc](/src/target/expect.go?s=4081:4178#L143)
``` go
func NewExpectProc(ctx context.Context, prog string, arg ...string) (*Expect, *os.Process, error)
```
NewExpectProc is similar to NewExpect except the created cmd is returned.
It is expected that the cmd will exit normally otherwise it is left to
the caller to kill it. This can be achieved by canceling the context @ctx.
In the event of an error starting the cmd it will be killed but not reaped.
However the cmd ends it is important that the caller reap the process
by calling cmd.Process.Wait() otherwise it can use up a process slot in
the operating system.
Note that Result is not filled in.


### <a name="Spawn">func</a> [Spawn](/src/target/expect.go?s=11281:11336#L402)
``` go
func Spawn(prog string, arg ...string) (*Expect, error)
```
Spawn is a wrapper to NewExpect(), for compatibility with the original expect





### <a name="Expect.BufStr">func</a> (\*Expect) [BufStr](/src/target/expect.go?s=10907:10941#L388)
``` go
func (exp *Expect) BufStr() string
```
BufStr is the buffer of expect read data as a string




### <a name="Expect.Clear">func</a> (\*Expect) [Clear](/src/target/expect.go?s=10799:10825#L383)
``` go
func (exp *Expect) Clear()
```
Clear out any unprocessed input




### <a name="Expect.Expect">func</a> (\*Expect) [Expect](/src/target/expect.go?s=6888:6959#L228)
``` go
func (exp *Expect) Expect(reOrStrs ...interface{}) (int, []byte, error)
```
Expect keeps reading input till either a timeout occurs (if set), one of the
strings/regexps passed matches the input, end of input occurs or an error.
If a string/regexp match occurs the index of the successful argument and the matching bytes
are returned. Otherwise an error value and error are returned.
Note: on EOF the return value will be NotFound and the error will be nil as
EOF is not considered an error. This is the only time those values will be returned.
See also Expecti()




### <a name="Expect.Expecti">func</a> (\*Expect) [Expecti](/src/target/expect.go?s=12342:12397#L437)
``` go
func (exp *Expect) Expecti(reOrStrs ...interface{}) int
```
Expecti is a convenience wrapper around Expect() that only returns the index
and not the found bytes or error. This is close to the original expect()




### <a name="Expect.Expectp">func</a> (\*Expect) [Expectp](/src/target/expect.go?s=12551:12606#L444)
``` go
func (exp *Expect) Expectp(reOrStrs ...interface{}) int
```
Expectp is a convenience wrapper around Expect() that panics on no match
(so will panic on eof)




### <a name="Expect.Kill">func</a> (\*Expect) [Kill](/src/target/expect.go?s=11045:11070#L393)
``` go
func (exp *Expect) Kill()
```
Kill the command. Using an Expect after a Kill is undefined




### <a name="Expect.LogUser">func</a> (\*Expect) [LogUser](/src/target/expect.go?s=12876:12911#L455)
``` go
func (exp *Expect) LogUser(on bool)
```
LogUser true asks for all input read by Expect() to be copied to stdout.
LogUser false (default) turns it off.
For compatibility with the original expect




### <a name="Expect.Send">func</a> (\*Expect) [Send](/src/target/expect.go?s=11458:11504#L407)
``` go
func (exp *Expect) Send(s string) (int, error)
```
Send sends the string to the process, for compatibility with the original expect




### <a name="Expect.SendL">func</a> (\*Expect) [SendL](/src/target/expect.go?s=12019:12066#L426)
``` go
func (exp *Expect) SendL(lines ...string) error
```
SendL is a convenience wrapper around Send(), adding linebreaks around each of the @lines.




### <a name="Expect.SendSlow">func</a> (\*Expect) [SendSlow](/src/target/expect.go?s=11671:11742#L413)
``` go
func (exp *Expect) SendSlow(delay time.Duration, s string) (int, error)
```
Sends string rune by rune with a delay before each.
Note: the return is the number of bytes sent not the number of runes sent




### <a name="Expect.SetCmdOut">func</a> (\*Expect) [SetCmdOut](/src/target/expect.go?s=5810:5856#L200)
``` go
func (exp *Expect) SetCmdOut(cmdOut io.Writer)
```
SetCmdOut if a non-nil io.Writer is passed it will be sent a copy of everything
read by Expect() from the pty.
Note that if you bypass expect and read directly from the *Expect this is
will not be used




### <a name="Expect.SetTimeout">func</a> (\*Expect) [SetTimeout](/src/target/expect.go?s=6132:6184#L212)
``` go
func (exp *Expect) SetTimeout(timeout time.Duration)
```
SetTimeout sets the timeout for future calls to Expect().
The default value is zero which cause Expect() to wait forever.




### <a name="Expect.SetTimeoutSecs">func</a> (\*Expect) [SetTimeoutSecs](/src/target/expect.go?s=6274:6320#L217)
``` go
func (exp *Expect) SetTimeoutSecs(timeout int)
```
SetTimeoutSecs is a convenience wrapper around SetTimeout




## <a name="ExpectWaitResult">type</a> [ExpectWaitResult](/src/target/expect.go?s=2885:2968#L114)
``` go
type ExpectWaitResult struct {
    ProcessState *os.ProcessState
    Error        error
}
```













- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
