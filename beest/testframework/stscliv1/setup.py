# -*- coding: utf-8 -*-

from setuptools import setup, find_packages

setup(
    name='stscliv1',
    version='0.1.0',
    description='StackState CLI wrapper',
    url='https://github.com/StackVista/beest/testframework/stscliv1',
    packages=find_packages(exclude=('tests')),
    install_requires=[
        "pytest-testinfra==6.6.0"
    ]
)
