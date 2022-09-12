from operator import truediv
import os
import sys

run_full_pipeline = True

for i in range(1, len(sys.argv)):
    if ".md" in str(sys.argv[i]) or ".markdown" in str(sys.argv[i]) or ".mdown" in str(sys.argv[i]):
        print("md file changed")
        run_full_pipeline = False
    else:
        run_full_pipeline = True
        break
    print('argument:', i, 'value:', sys.argv[i])

print(run_full_pipeline)

if run_full_pipeline:
    with open('.env', 'w') as writer:
        writer.write(f'export RUN_FULL_PIPELINE="false"')
else:
    with open('.env', 'w') as writer:
        writer.write(f'export RUN_FULL_PIPELINE="false"')