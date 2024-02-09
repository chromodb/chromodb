#!/bin/bash

#
# ChromoDB
###########################################################################
# Originally authored by Alex Gaetano Padula
# Copyright (C) ChromoDB
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


VERSION=v0.9.3

echo "🛠️ Building ChromoDB $VERSION multiplatform binaries!"

( GOOS=darwin GOARCH=amd64 go build -o bin/macos-darwin/amd64/chromodb && tar -czf bin/macos-darwin/amd64/chromodb-$VERSION-amd64.tar.gz -C bin/macos-darwin/amd64/ $(ls  bin/macos-darwin/amd64/))
( GOOS=darwin GOARCH=arm64 go build -o bin/macos-darwin/arm64/chromodb && tar -czf bin/macos-darwin/arm64/chromodb-$VERSION-arm64.tar.gz -C bin/macos-darwin/arm64/ $(ls  bin/macos-darwin/arm64/))
( GOOS=linux GOARCH=386 go build -o bin/linux/386/chromodb && tar -czf bin/linux/386/chromodb-$VERSION-386.tar.gz -C bin/linux/386/ $(ls  bin/linux/386/))
( GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/chromodb && tar -czf bin/linux/amd64/chromodb-$VERSION-amd64.tar.gz -C bin/linux/amd64/ $(ls  bin/linux/amd64/))
( GOOS=linux GOARCH=arm go build -o bin/linux/arm/chromodb && tar -czf bin/linux/arm/chromodb-$VERSION-arm.tar.gz -C bin/linux/arm/ $(ls  bin/linux/arm/))
( GOOS=linux GOARCH=arm64 go build -o bin/linux/arm64/chromodb && tar -czf bin/linux/arm64/chromodb-$VERSION-arm64.tar.gz -C bin/linux/arm64/ $(ls  bin/linux/arm64/))
( GOOS=freebsd GOARCH=arm go build -o bin/freebsd/arm/chromodb && tar -czf bin/freebsd/arm/chromodb-$VERSION-arm.tar.gz -C bin/freebsd/arm/ $(ls  bin/freebsd/arm/))
( GOOS=freebsd GOARCH=amd64 go build -o bin/freebsd/amd64/chromodb && tar -czf bin/freebsd/amd64/chromodb-$VERSION-amd64.tar.gz -C bin/freebsd/amd64/ $(ls  bin/freebsd/amd64/))
( GOOS=freebsd GOARCH=386 go build -o bin/freebsd/386/chromodb && tar -czf bin/freebsd/386/chromodb-$VERSION-386.tar.gz -C bin/freebsd/386/ $(ls  bin/freebsd/386/))
( GOOS=windows GOARCH=amd64 go build -o bin/windows/amd64/chromodb.exe && zip -r -j bin/windows/amd64/chromodb-$VERSION-x64.zip bin/windows/amd64/chromodb.exe)
( GOOS=windows GOARCH=arm64 go build -o bin/windows/arm64/chromodb.exe && zip -r -j bin/windows/arm64/chromodb-$VERSION-x64.zip bin/windows/arm64/chromodb.exe)
( GOOS=windows GOARCH=386 go build -o bin/windows/386/chromodb.exe && zip -r -j bin/windows/386/chromodb-$VERSION-x86.zip bin/windows/386/chromodb.exe)


echo "✅ Fin.  Binaries are available under ./bin directory."