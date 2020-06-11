from tika import parser
import sys

fname = sys.argv[1]

file = open(fname + ".temp", "w+")

raw = parser.from_file(sys.argv[1])

content = raw['content']

for i in content:
    try:
        file.write(i)
    except:
        pass