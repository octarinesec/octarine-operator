#!/usr/bin/env python
import os
import sys

if len(sys.argv) > 1:
    deploy_dir = sys.argv[1]
else:
    deploy_dir = '../deploy'

namespace_file = 'namespace.yaml'
namespace = 'octarine-dataplane'
output_file = 'setup.yaml'

yaml_list = [namespace_file]
files = []

# collect all yaml file names
for dir_name, subdir_list, file_list in os.walk(deploy_dir):
    for file_name in file_list:
        if file_name.endswith('.yaml'):
            yaml_list.append('{}/{}'.format(dir_name, file_name))

# collect all yaml files text
for file_name in yaml_list:
    with open(file_name) as file:
        file_text = ''
        for line in file.readlines():
            # add namespace to non-cluster entities
            if not (file_name == namespace_file or file_name.startswith('cluster')):
                if line == 'metadata:\n':
                    line = 'metadata:\n  namespace: {}\n'.format(namespace)
            file_text += line
        files.append(file_text)

# concat and output to file
output = "\n---\n\n".join(files)
with open(output_file, 'w+') as file:
    file.writelines(output)
