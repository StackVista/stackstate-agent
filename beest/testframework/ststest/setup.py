# -*- coding: utf-8 -*-

from setuptools import setup, find_packages


with open('README.md') as f:
    readme = f.read()

setup(
    name='ststest',
    version='0.1.0',
    description='A bunch of code to test StackState data',
    long_description=readme,
    url='https://github.com/StackVista/beest/testframework/ststest',
    license=license,
    packages=find_packages(exclude=('tests'))
)
