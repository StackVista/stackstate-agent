# Vagrant Assistance

This vagrant image can be used to create the required env to run the following scenarios:

- rtloader.make
- rtloader.test

Run the `up.sh` to setup the required env.

When that is complete then run any of the following scripts:

- rtloader.test.sh
  - This will run `rtloader.make` and `rtloader.test` within this project

Feel free to add more `config.vm.provision` scenarios below rtloader.test to create more test cases
