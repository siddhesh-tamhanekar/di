## Compile Time Dependancy Injection for Golang

This library generates constructor functions with dependancy code for struct and interfaces.

The library is inspired by google [wire](https://github.com/google/wire)
#### Getting Started

Download and install using following commands

`go get github.com/siddhesh-tamhanekar/di`

`go install github.com/siddhesh-tamhanekar/di`


Execute following command for generating dependancies in your project.

`<goroot>/bin/di`

#### Running Example from source code

`go run .`


---

#### How to use the Library

- Create file called di.go (this file can be declared in every package where as per your need)
- `go:build exclude` to ignore di.go while compiling project.
- This file will contain the code which tells library about how the dependancies should be resolved.
- we can use the library methods (mentioned below) to declare the ependancies.
- Once di.go is ready we can run `<goroot>/bin/di.go` to generate dependancies.
- Refer example directory for more details.

#### Methods
| Function   | Usage   |
| ------------ | ------------ |
|  Share  | `di.Share(yourStruct{},db)`<br> the second parameter is the package level variable which need to use while resolving dependancy it will act like a signleton|
|  Build |  `di.Build(yourStruct{})` <br> build method will create constructor function for given struct|
|  Bind | `di.Bind(yourInterface, targetStruct{}`<br> this will bind Interface to struct  |


#### Conventions
All the functions generated with `Build` method call will start with `New` keyword followed by struct name for e.g. generated function for `UserRepository` struct will be `NewUserRepository() UserRepository`

All the functions genertated with `Bind` method call will start with `New` keyword followed by interface name for e.g. generated function for `UserCreator` interface will be `New UserCreator() UserCreator`

#### Closing Thoughts
This library is still in beta and needs to handle the edge case scenerios. pls feel free to open issues and pull request to enrich the library.

