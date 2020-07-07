# Response hash
## What it does
It generates md5 hash from response body of URLs listed in input file and saves them to a file.

It can be helpful when refactoring (GET) APIs to make sure that the response doesn't change. At least that's what I made it for. 
1.) Generate hashes before changes
2.) Generate hashes after changes
3.) Throw in some online comparison tool :D

## How to use
Write URLs you want to generate hashes for into a file, separated by new line. Then run the program => point it to a file (relative to where it's executed). The rest is self explanatory