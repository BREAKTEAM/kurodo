# kurodo

kurodo is a web directory/file finder and a HTTP request fuzzer.
An occurence of the `FUZZ` keyword (anywhere in the request) will be replaced by an entry from the wordlist.

## E.g

```bash
$ kurodo -u domain.com -w wordlists/wordlists.txt

kurodo Fuzzy Tools By Aishee

---------------------------------------------------------------------------------
Chars(-hh)    Words(-hw)   Lines(-hl)   Header(-hr)  Code(-hc)    Result
---------------------------------------------------------------------------------
185           22           7            140          301          Admin
185           22           7            140          301          Login
185           22           7            140          301          login
0             0            0            198          200          passwords
185           22           7            119          301          test
```

## Install

Install Go before:
Link tutorial for ubuntu: <https://tecadmin.net/install-go-on-ubuntu/https://tecadmin.net/install-go-on-ubuntu/>

Install kurodo:

```bash
go get github.com/BREAKTEAM/kurodo
cd $GOPATH/src/github.com/BREAKTEAM/kurodo
go install
kurodo -h
```

## Usage

Find hidden files:

```bash
kurodo -u domain.com -w wordlists.txt
```

Brute force a header field:

```bash
kurodo -u domain.com -w wordlists.txt -H "User-Agent: Kurodo"
```

Brute force a file extension:

```bash
kurodo -u domain.com/file.Kurodo -w ext.txt
```

Brute force a password send via a form with POST:

```bash
kurodo -u domain.com/login.php -w wordlists.txt -m POST \
    -d "user=admin&passwd=Kurodo&submit=s" \
    -H "Content-Type: application/x-www-form-urlencoded"
```

Brute force HTTP methods:

```bash
kurodo -u domain.com -w wordlists.txt -m Kurodo
```

## Docker

Build the image:

```bash
cd $GOPATH/src/github.com/BREAKTEAM/kurodo
docker build -t kurodo .
```

Run kurodo with docker:

```bash
docker run -v $(pwd)/wordlists:/wordlists kurodo -u domain.com -w /wordlists/wordlists.txt
```

## Wordlists recommended

[SecLists](https://github.com/danielmiessler/SecLists/tree/master/Discovery/Web-Content)
[Crackstation](https://crackstation.net/crackstation-wordlist-password-cracking-dictionary.htm)
<https://github.com/berzerk0/Probable-Wordlists>
