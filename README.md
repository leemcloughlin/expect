
# expect
    import "github.com/leemcloughlin/expect"

Expect is pure Go (golang) version of the terminal interaction package Expect
common on many Linux systems

A very simple example that calls the rev program to revese the text in each
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




## Constants
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

## Variables
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


## type Expect
``` go
type Expect struct {
    // *os.File is an anonymous field for the pty connected to the command.
    // This allows you to treat *Expect as a *os.File
    *os.File

    // Expect reads into here. On a successful match all data up the end of the
    // match is deleted shrinking the buffer.
    // See also Clear() and BufStr()
    Buffer *bytes.Buffer

    // On EOF being read from Cmd this is set (and ExpectReader is ended)
    Eof bool
    // contains filtered or unexported fields
}
```








### func NewExpect
``` go
func NewExpect(prog string, arg ...string) (*Expect, error)
```
NewExpect starts prog, passing any given args, in its own pty.
Note that in order to be non-blocking while reading from the pty this sets
the non-blocking flag and looks for EAGAIN on reads failing.  This has only
been tested on Linux systems


### func Spawn
``` go
func Spawn(prog string, arg ...string) (*Expect, error)
```
Spawn is a wrapper to NewExpect(), for compatibility with the original expect




### func (\*Expect) BufStr
``` go
func (exp *Expect) BufStr() string
```
BufStr is the buffer of expect read data as a string



### func (\*Expect) Clear
``` go
func (exp *Expect) Clear()
```
Clear out any unprocessed input



### func (\*Expect) Expect
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



### func (\*Expect) Expecti
``` go
func (exp *Expect) Expecti(reOrStrs ...interface{}) int
```
Expecti is a convenience wrapper around Expect() that only returns the index
and not the found bytes or error. This is close to the original expect()



### func (\*Expect) Expectp
``` go
func (exp *Expect) Expectp(reOrStrs ...interface{}) int
```
Expectp is a convenience wrapper around Expect() that panics on no match
(so will panic on eof)



### func (\*Expect) Kill
``` go
func (exp *Expect) Kill() error
```
Kill the command. Using an Expect after a Kill is undefined



### func (\*Expect) LogUser
``` go
func (exp *Expect) LogUser(on bool)
```
LogUser true asks for all input read by Expect() to be copied to stdout.
LogUser false (default) turns it off.
For compatibility with the original expect



### func (\*Expect) Send
``` go
func (exp *Expect) Send(s string) (int, error)
```
Send sends the string to the process, for compatibility with the original expect



### func (\*Expect) SendSlow
``` go
func (exp *Expect) SendSlow(delay time.Duration, s string) (int, error)
```
Sends string rune by rune with a delay before each.
Note: the return is the number of bytes sent not the number of runes sent



### func (\*Expect) SetCmdOut
``` go
func (exp *Expect) SetCmdOut(cmdOut io.Writer)
```
SetCmdOut if a non-nil io.Writer is passed it will be sent a copy of everything
read by Expect() from the pty.
Note that if you bypass expect and read directly from the *Expect this is
will not be used



### func (\*Expect) SetTimeout
``` go
func (exp *Expect) SetTimeout(timeout time.Duration)
```
SetTimeout sets the timeout for future calls to Expect().
The default value is zero which cause Expect() to wait forever.



### func (\*Expect) SetTimeoutSecs
``` go
func (exp *Expect) SetTimeoutSecs(timeout int)
```
SetTimeoutSecs is a convenience wrapper around SetTimeout









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)