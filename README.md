# Terraform Targeter
A command line tool that helps you construct targeted apply commands by letting you interactively select targets from your plan.

## Installation
Installing `tf-targeter` requires that you have `go` installed on your system. You can find downloads for `go` on their website: https://go.dev/
```
git clone https://github.com/audunmo/tf-targeter
cd tf-targeter
go install
```

## Usage
Run `tf-targeter` in any folder that contains a terraform config. If there's a planned change, the relevant resources will be presented for you to choose from. When you've made your selection, `tf-targeter` will output the right command. 
