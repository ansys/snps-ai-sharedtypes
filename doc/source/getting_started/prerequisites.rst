Prerequisites
=============

Before working with AALI Shared Types, you need Go and Git installed on your system.

Go Installation
---------------

AALI Shared Types requires **Go 1.23.0 or higher**.

.. tab-set::

    .. tab-item:: macOS

        Using Homebrew:

        .. code:: bash

           brew install go

        Or download from `golang.org <https://golang.org/doc/install>`_

    .. tab-item:: Windows

        1. Download Go from `golang.org <https://golang.org/doc/install>`_
        2. Run the MSI installer
        3. The installer will add Go to your PATH automatically

    .. tab-item:: Linux

        .. code:: bash

           # Ubuntu/Debian
           sudo apt update
           sudo apt install golang-go

           # Or download the latest version from golang.org

Git Installation
----------------

Git is required to fetch Go modules and clone the repository.

.. tab-set::

    .. tab-item:: macOS

        .. code:: bash

           brew install git

    .. tab-item:: Windows

        Download from `git-scm.com <https://git-scm.com/download/win>`_

    .. tab-item:: Linux

        .. code:: bash

           sudo apt install git

Clone the Repository
--------------------

To work on AALI Shared Types:

.. code:: bash

   git clone https://github.com/ansys/snps-ai-sharedtypes.git
   cd snps-ai-sharedtypes

Verify Installation
-------------------

Check that everything is installed correctly:

.. code:: bash

   go version  # Should show 1.23.0 or higher
   git --version

Next Steps
----------

With prerequisites installed, you can:

- :doc:`Install <installation>` AALI Shared Types as a dependency
- :doc:`Add custom types <adding_custom_types>` for your FlowKit functions
