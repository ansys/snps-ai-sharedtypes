Installation
============

AALI Shared Types is a Go module that provides type definitions for AALI services.

Import as Go Module
-------------------

To use AALI Shared Types in your Go project, import it as a module dependency:

.. code-block:: bash

   go get github.com/ansys/snps-ai-sharedtypes

Using in Your Code
------------------

Import the packages you need:

.. code-block:: go

   import (
       "github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
       "github.com/ansys/snps-ai-sharedtypes/pkg/typeconverters"
   )

Verify Installation
-------------------

Create a simple test file to verify the import works:

.. code-block:: go

   package main

   import (
       "fmt"
       "github.com/ansys/snps-ai-sharedtypes/pkg/sharedtypes"
   )

   func main() {
       // Create a function definition
       funcDef := sharedtypes.FunctionDefinition{
           Name: "test-function",
           Type: "go",
       }

       fmt.Printf("Created function: %s\n", funcDef.Name)
   }

Run the test:

.. code-block:: bash

   go run main.go

Next Steps
----------

Now that you have AALI Shared Types installed, learn how to :doc:`add custom types <adding_custom_types>` for your FlowKit functions.
