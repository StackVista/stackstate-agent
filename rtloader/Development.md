# Agent Python3 build
```shell
conda create -n ddpy3 python python=3.8

conda activate ddpy3
pip install invoke distro==1.4.0 awscli
inv deps

inv rtloader.clean && inv rtloader.make --python-runtimes 3 && inv rtloader.test
```

# Agent Python2 build
```shell
conda create -n ddpy2 python python=2

conda activate ddpy2
pip install invoke distro==1.4.0 awscli
inv deps

inv rtloader.clean && inv rtloader.make --python-runtimes 2
inv rtloader.test
```

