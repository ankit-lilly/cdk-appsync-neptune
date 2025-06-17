jsii => Framework that aws build that allows you to use cdk with languages that aren't typescripit.

If we are using TS then we can use CDK directly but with Go we need to use jsii to convert the code to typescript and then use CDK.
# CDK Notes

-> We create an app.
-> We create an stack and pass our app to it.

-> For any resources we create we pass our stack to it so this way all the resources are part of the stack.




## CDK with Go

- Since the CDK is written in TypeScript, we need to wrap our Go data types in jsii to use them in CDK.

- Go is not a native runtime for CDK, so we need to use ProvidedRuntime.

But this is not a problem because Go doesn't need a runtime like Node or Python.
And since it support cross-compilation, we can compile it and package it for the lambda function.

The way it works is that we create a binary, zip it and specify the path via Code property using lambda.AssetCode_FromAsset

Note: the binary has to be named bootstrap otherwise lambda  won't be able to find it and it has to be compiled to linux/amd64 architecture.

The handler will always be main.




