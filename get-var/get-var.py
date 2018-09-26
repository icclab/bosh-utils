#!/usr/bin/env python3

import subprocess
from pyperclip import copy
from os import path, devnull
from sys import argv
from colorama import Fore, Style

# Help
if len(argv) == 1 or argv[1] == '-h' or 'help' in argv[1]:
    print("Usage: [VAR_NAME]\nUsage: [VAR_NAME] -v (Gives details about retrieved variable\nUsage: grep [SEARCH_TERM]")
    print("Note: use 'paths' to get the current var files being searched")
    exit()

# Paths for vars files
null_file = open(devnull, 'w')
var_files=["~/Desktop/active-deployment/bosh/vars/director-vars-store.yml", "~/Desktop/active-deployment/cf-deployment/custom/cf-vars-store.yml", 
    "~/Desktop/active-deployment/cf-deployment/custom/cf-vars-file.yml", "~/Desktop/active-deployment/services/cf-mysql-deployment/custom/cf-mysql-vars.yml", 
    "~/Desktop/active-deployment/services/cf-rabbitmq-multitenant-broker-release/custom/mtrmq-vars-store.yml", 
    "~/Desktop/active-deployment/services/cf-rabbitmq-multitenant-broker-release/custom/mtrmq-vars-file.yml",
    "~/Desktop/active-deployment/concourse/concourse-vars.yml", "~/Desktop/active-deployment/services/prometheus-boshrelease/custom/prom-vars-store.yml"]

if argv[1] == 'paths':
    for f in var_files:
        print(f)
    exit()

# Expand path
for i in range(len(var_files)):
    var_files[i] = path.expanduser(var_files[i])

# Try to get a specific variable
def get_var(f):
    var = Fore.RED + argv[1] + Style.RESET_ALL
    try:
        output = str(subprocess.check_output(["bosh", "int", f, "--path", "/" + argv[1]], stderr=null_file, encoding='utf-8'))
        if '-v' in argv:
            print(var + ":", output.strip(), "(" + "found in", "'" + Fore.YELLOW + f + Style.RESET_ALL + "')")

        print(var, 'copied to clipboard')        
        copy(output.strip())
        return True
    except subprocess.CalledProcessError:
            return False

# Grep all files
def grep_var(f):
    try:
        output = str(subprocess.check_output(["grep", argv[2], f]), encoding = 'utf-8')
        print(Fore.BLUE + path.basename(f) + Style.RESET_ALL)

        # Color var names and print lines
        lines = output.splitlines()
        for l in lines:
            split = l.split(':')

            if len(split) == 2:
                print("  " + Fore.RED + split[0] + Style.RESET_ALL + ":" + split[1])
            else:
                print(l)

        print('')
    except subprocess.CalledProcessError:
        pass

found = False
for f in var_files:
    if not path.isfile(f):
        print("'" + f + "'", 'does not exist or is not a file')
        continue
    
    print(Style.RESET_ALL, end='')
    
    if argv[1] == 'grep' and len(argv) > 2:
        grep_var(f)
        found = True
    elif get_var(f) == True:
        found = True
        break

if found == False:
    print(Fore.RED + argv[1] + Style.RESET_ALL, "not found!")
