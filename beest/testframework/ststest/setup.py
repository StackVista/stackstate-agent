# -*- coding: utf-8 -*-

from setuptools import setup, find_packages

setup(
    name='ststest',
    version='0.1.0',
    description='A bunch of code to test StackState data',
    url='https://github.com/StackVista/beest/testframework/ststest',
    packages=find_packages(exclude=('tests')),
    install_requires=[
        "pytest-testinfra==6.6.0",
        "marshmallow==3.17.0",
        "stscliv1",
        "pydot",
        "pyshorteners"
    ]
)
