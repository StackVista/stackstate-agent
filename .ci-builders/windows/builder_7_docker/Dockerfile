FROM vinogradoff/windows-ruby:2.4.3.2 as base

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
ARG GO_VERSION="1.18.0"
WORKDIR C:/

ADD extras .
#ADD extras/scripts .
#ADD extras/binaries C:/Windows/Temp
#ADD extras/files C:/files/
#ADD extras/sxs C:/sxs/
#RUN Write-Host "Go Version: $env:GO_VERSION"

FROM base as intermediate_1

ARG GO_VERSION

RUN C:/scripts/not_all_in_one.ps1 -goVersion $env:GO_VERSION
RUN C:/scripts/all_in_one.ps1
#
##RUN Write-Host "Go Version: $env:GO_VERSION"
#
#FROM intermediate_1 as intermediate_2
#ARG GO_VERSION
#
#RUN C:/choco.ps1 -goVersion $env:GO_VERSION
#
#FROM intermediate_2 as intermediate_3
#
#RUN C:/provisioner.ps1
#
#FROM intermediate_3 as intermediate_4
#
#RUN C:/choco_part2.ps1
#
#FROM intermediate_4 as intermediate_5
#
#RUN C:/choco_part3.ps1
#
#FROM intermediate_5 as intermediate_6
#
#RUN C:/choco_part4.ps1
#
#FROM intermediate_6 as intermediate_7
#
#RUN C:/provisioner_part2.ps1
#
#FROM intermediate_7 as intermediate_8
#
#RUN C:/choco_part5.ps1
