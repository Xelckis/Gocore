# Gocore
Gocore is a project that provides a robust, reliable implementation of some of the standard POSIX core utilities, written entirely in the Go programming language.
# Commands
- comm
- head
- more
- tail
- cat
- cp
- ln
- mv
- tee
- chown
- cut
- touch
- cmp
- mkdir
- rm
- uniq
- ls
- cal

# Comm
The comm utility shall read file1 and file2, which should be ordered in the current collating sequence, and produce three text columns as output: lines only in file1, lines only in file2, and lines in both files.

If the lines in both files are not ordered according to the collating sequence of the current locale, the results are unspecified.

## Usage and flags
```comm [−123] file1 file2```

The following options are supported:

- ```−1```  Suppress the output column of lines unique to file1.
- ```−2```  Suppress the output column of lines unique to file2.
- ```−3```  Suppress the output column of lines duplicated in file1 and file2.

# Head
The head utility shall copy its input files to the standard output, ending the output for each file at a designated point. Copying shall end at the point in each input file indicated by the −n number option.

## Usage and flags
```head [−n number] [file...]```

The following option are supported:

- ```n, --lines```  The first number lines of each input file shall be copied to standard output.

# More
The more utility shall read files and either write them to the terminal on a page-by-page basis or filter them to standard output.

## Usage and flags
```more [−cis] [−n number] [−p command] [−t tagstring] [file...]```

The following options are supported:

- ```-c, --clean-print```  Do not scroll. Instead, paint each screen from the top, clearing the remainder of each line as it is displayed.
- ```-i, --case-insensitive```   Perform pattern matching in searches without regard to case
- ```-p, --command string```  Each time a screen from a new file is displayed or redisplayed, execute the more command(s) in the command arguments in the order specified, as if entered by the user after the first screen has been displayed.
- ```-n, --lines int```  Specify the number of lines per screenful. The number argument is a positive decimal integer. The --lines option shall override any values obtained from any other source, such as number of lines reported by terminal.
- ```-s, --squeeze```  Squeeze multiple blank lines into one.
- ```-t, --tag string```  Start displaying the file from the first line containing the specified tag. If the tag is not found, display begins from the start of the file.

# Tail
The tail utility shall copy its input file to the standard output beginning at a designated place. Tails is relative to the end of the file.

## Usage and flags
```tail [−f] [−c number|−n number] [file]```

The following options are supported:

- ```-c, --bytes string```  Output the last NUM bytes; or use -c +NUM to output starting with byte NUM of each file
- ```-f, --follow```  Output appended data as the file grows;
- ```-n, --lines string ```  Output the last NUM lines, instead of the last 10; or use -n +NUM to skip NUM-1 lines at the start

# Cat 
The cat utility shall read files in sequence and shall write their contents to the standard output in the same sequence.

## Usage and flags
```cat [−u] [file...]```

The following options are supported:

- ```-u, --bytes```   Write bytes from the input file to the standard output without delay as each is read.

# Cp 
The cp utility shall copy the contents of source_file to the destination path named by target.

## Usage and flags
```
cp [−Pp] source_file target_file
cp [−Pp] source_file... target
cp −R [−H|−L|−P] [−fip] source_file... target
```

The following options are supported:

- ```-L, --dereference```  Always follow symbolic links in SOURCE
- ```-H, --follow-symbolic```  Follow command-line symbolic links in SOURCE
- ```-P, --no-dereference```  Never follow symbolic links in SOURCE
- ```-p, --preserve```  Preserve the file attributes
- ```-r, --recursive ```  Copy directories recursively

# Ln
the ln utility shall create a new directory entry (link) at the destination path specified by the target_file operand.

## Usage and flags
```
ln [−fs] [−L|−P] source_file target_file
ln [−fs] [−L|−P] source_file... target_dir
```

The following options are supported:

- ```-f, --force``` Remove existing destination files
- ```-L, --logical```  Dereference TARGETs that are symbolic links
- ```-P, --physical```  Make hard links directly to symbolic links
- ```-s, --symbolic ``` Make symbolic links instead of hard links

# Mv
the mv utility shall move the file named by the source_file operand to the destination specified by the target.

## Usage and flags
```
mv [−if] source_file target_file
mv [−if] source_file... target_dir
```

The following options are supported:

- ```-f, --force```  Do not prompt before overwriting
- ```-i, --interactive```  Prompt before overwrite

# Tee
The tee utility shall copy standard input to standard output, making a copy in zero or more files.

## Usage and flags
```
tee [−ai] [file...]
```

The following options are supported:

- ```-a, --append``` Append to the given FILEs, do not overwrite
- ```-i, --ignore-interrupts``` Ignore interrupt signals

# Chown
The chown utility shall set the user ID of the file named by each file operand to the user ID specified by the owner operand.

## Usage and flags
```
chown [−d] owner[:group] file...
chown −R [−H|−L|−P] owner[:group] file...
```

The following options are supported:

- ```-d, --no-dereference```  Affect symbolic links instead of any referenced file (useful only on systems that can change the ownership of a symlink)
- ```-R, --reccursive```  Operate on files and directories recursively
- ```-P, --physical```  Do not traverse any symbolic links
- ```-H, --Hybrid```  If a command line argument is a symbolic link to a directory, traverse it
- ```-L, --logical```   Traverse every symbolic link to a directory encountered

# Cut
The cut utility shall cut out bytes, characters, or character-delimited fields from each line in one or more files.

## Usage and flags
```
cut −c list [file...]
cut −f list [−d delim] [−s] [file...]
```

The following options are supported:

- ```-c, --characters string```  Select only these characters
- ```-d, --delimiter string```  Use DELIM instead of TAB for field delimiter
- ```-f, --fields string```  Select only these fields;  also print any line that contains no delimiter character, unless the -s option is specified
- ```-s, --only-delimited```  Do not print lines not containing delimiters

# Touch
The touch utility shall change the last data modification timestamps, the last data access timestamps, or both.

## Usage and flags
```
touch [−acm] [−r ref_file|−t time|−d date_time] file...
```

The following options are supported:

- ```-a, --access```  Change only the access time
- ```-d, --date string```  Parse STRING and use it instead of current time
- ```-m, --modify```  Change only the modification time
- ```-c, --no-create```  Do not create any files
- ```-r, --reference```  Use this file's times instead of current time
- ```-t, --timestamp string```  Use specified time instead of current time, with a date-time format that differs from -d's

# Cmp
The cmp utility shall compare two files. The cmp utility shall write no output if the files are the same. Under default options, if they differ, it shall write to standard output the byte and line number at which the first difference occurred.

## Usage and flags
```
cmp [−l|−s] file1 file2
```

The following options are supported:
- ```-s, --quiet``` Suppress all normal output
- ```-l, --verbose```  Output byte numbers and differing byte values

# Mkdir
The mkdir utility shall create the directories specified by the operands, in the order specified.

## Usage and flags
```
mkdir [−p] [−m mode] dir...
```

The following options are supported:

- ```-m, --mode``` Set file mode (default 0755)
- ```-p, --parents```  Create any missing intermediate pathname components.

# Rm
The rm utility shall remove the directory entry specified by each file argument.

## Usage and flags
```
rm [−ir] file...
rm −f [−ir] [file...]
```

The following options are supported:

- ```-f, --force```  Do not prompt for confirmation. Do not write diagnostic messages or modify the exit status in the case of no file operands, or in the case of operands that do not exist.
- ```-i, --interactive```  Prompt before every removal
- ```-r, --recursive```  Remove file hierarchies.
         
# Uniq
The uniq utility shall read an input file comparing adjacent lines, and write one copy of each input line on the output. The second and succeeding copies of repeated adjacent input lines shall not be written.

## Usage and flags
```
uniq [−c|−d|−u] [−f fields] [−s char] [input_file [output_file]]
```

The following options are supported:

- ```-c, --count```        Prefix lines by the number of occurrences
- ```-d, --repeated```     Only print duplicate lines, one for each group
- ```-s, --skip-chars```   Avoid comparing the first N characters
- ```-f, --skip-fields```  Avoid comparing the first N fields
- ```-u, --unique ```      Only print unique lines

# Ls
For each operand that names a file of a type other than directory or symbolic link to a directory, ls shall write the name of the file as well as any requested, associated information.

## Usage and flags
```
ls [−ikqr] [−g lno ] [−A|−a] [−C|−m|−1] [−F|−p] [−L] [−R|−d] [−S|−f|−t] [−c|−u] [file...]
```

The following options are supported:

- ```-a, --all```                  Do not ignore entries starting with .
- ```-A, --almost-all```           Lists all entries, including hidden files (those starting with a .), but excludes the current directory (.) and the parent directory (..).
- ```-c, --change-time```          Use time of last modification of the file status information
- ```-F, --classify```             This flag appends a character to the end of each filename to indicate its type (/*@|).
- ```-C, --column```               Forces the output into multiple columns
- ```-L, --dereference```          When showing file information for a symbolic link, show information for the file the link references rather than for the link itself
- ```-q, --hide-control-chars```   Print ? instead of nongraphic characters
- ```-p, --indicator-style```      Append / indicator to directories
- ```-k, --kibibytes```            Default to 1024-byte blocks for file system usage; used only with -s and per directory totals
- ```-l, --long-listing```         Use a long listing format
- ```-n, --numeric-uid-gid```      Turn on the −l (ell) option, but when writing the file’s owner or group, write the file’s numeric UID or GID rather than the user or group name.
- ```-o, --omit-group```           Like -l, but do not list group information
- ```-g, --omit-owner```           Like -l, but do not list owner
- ```-1, --one-per-line```         List one file per line
- ```-R, --recursive```            List subdirectories recursively
- ```-r, --reverse-sort```         Reverse order while sorting
- ```-i, --show-inode```           For each file, write the file’s file serial number (inode)
- ```-S, --size```                 Sort by file size, largest first
- ```-t, --sort-mtime```           Sort by time, newest first
- ```-m, --stream-format```        Fill width with a comma separated list of entries
- ```-u, --access-time```          Use time of last access instead of last modification of the file for sorting (−t) or writing (−l).
- ```-f, --no-sort```              Do not sort

# Cal
The cal utility shall write a calendar to standard output using the Gregorian calendar

## Usage
```cal [[month] year]```
