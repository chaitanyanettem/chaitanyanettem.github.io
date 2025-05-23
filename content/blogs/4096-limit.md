{
    "date": "2015-12-06",
    "tags": ["shell", "quicktip", "linux"],
    "short": "An interesting tidbit about stdin and Linux",
    "slug": "4096-limit"
}

# Why there's a 4096 character limit on input from stdin

*Originally posted at http://blog.chaitanya.im/4096-limit*

The other day I had just started working on puzzles in [Advent of Code](https://adventofcode.com/2015) after seeing it on HackerNews (The puzzles are fun. You should definitely try them out). The first half of the first problem was very easy and so I finished it in less than 5 minutes. Only to find out that my program was generating the wrong solution for the given input. I was confounded. There was nothing logically wrong with it. [This](https://adventofcode.com/2015/day/1) was the problem and the solution looked like this:

```python
class day_1:
    str = ""
    floor = 0
    def __init__(self, str):
        self.str = str

    def calculate(self):
        for c in str:
            if c == '(':
                self.floor += 1
            else:
                self.floor -= 1

if __name__ == "__main__":
    str = input("Enter input string: ")
    print(len(str))
    x = day_1(str)
    x.calculate()
    print(x.floor)
```

As you can see the program is pretty simple; it reads an input string from the commandline and processes the string. On checking the size of the input string in the program, I found that its size was 4095 characters. Comparing that to the actual input string from the website I found that *its* size was some 7000 odd characters. The reason for this discrepancy as I later found out is that the Linux commandline (and I believe all POSIX compliant commandlines) can take upto 4096 characters of input (4kB) only.

If you just want to know what the solution to this problem is then here it is. There are two options:
- Read from a file instead of taking input from the command line.
- Or enter the command `stty -icanon` before starting your program. This will change the input mode to `noncanonical` mode which allows your inputs to be as big as you want. Be aware that this will remove all text processing capabilities from commandline input. I.e. you won’t be able to use your arrow keys, backspace, delete, ability to end input using ctrl-D etc. `stty icannon` will revert the input to canonical mode.

If you’re now wondering what canonical and non-canonical mean then read on.

## What is canonical mode? Why does it limit the input size?

How the data input in the terminal by the user is provided to a process depends on whether the terminal is in the canonical or non-canonical mode.

- In Canonical mode the terminal provides basic text processing abilities on input: You can navigate inside the input buffer using arrow keys; erase, delete, EOF signals are captured and applicable text processing is performed. Text is passed to the process only when a new line is encountered or when the buffer is full. The size of the buffer is usually 4096 and **this** is why the input size is limited. As mentioned above you can switch from canonical to non-canonical mode using `stty -icanon`.
- Non-Canonical mode has none of these text-processing capabilities. I.e if you wanted to implement a commandline tool like `cat` and you wanted it to operate in a non-canonical environment you’d have to implement some of these capabilities yourself within your program. You can switch back to canonical mode using `stty icanon`.

More info is available here:

1. [The Open Group Base Specifications Issue 7 — General Terminal Interface](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap11.html#tag_11_01_05)
2. [This unix.stackexchange question](https://unix.stackexchange.com/questions/131105/how-to-read-over-4k-input-without-new-lines-on-a-terminal)
3. Also, you can see the actual buffer size defined [here](https://github.com/torvalds/linux/blob/3088d26962e802efa3aa5188f88f82a957f50b22/drivers/tty/n_tty.c#L59) (`N_TTY_BUF_SIZE` on line #59)