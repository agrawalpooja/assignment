CS733 Assignment 1

Submitted By :
Roll No : 153050035
Name : Pooja Agrawal

1) Client can send 4 different commands(write,read,cas,delete) to server and will get corresponding output as specified in the assignment.

2) It is assumed that server is always up (as there is no persistence storage of file information, which was allowed).

3) If expiry time is not given in the command then it is taken as 9999999999 (assuming it as infinity as it is a quite bigger number)

4) Expiry time is assumed in seconds.

5) Different errors handled :
	-> ERR_FILE_NOT_FOUND\r\n
	-> ERR_CMD_ERR\r\n
	-> ERR_FILE_EXPIRED\r\n
	-> ERR_VERSION <newversion>\r\n 
	-> ERR_INTERNAL\r\n  (if data content in command exceeds than the number of bytes specified, and in that case file write operation returns with this error)

6) Multiple commands can be present within same command.
	e.g. write abc.txt 10 100\r\nhelloworld\r\nreadabc.txt\r\n
	