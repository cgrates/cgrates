Troubleshooting
===============

In case of **troubleshooting**, CGRateS can monitor memory profiling and CPU profiling.

Memory Profiling
----------------

Creating the memory profile files
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Firstly, go to the main config directory and run the engine with the flag named ```-memprof_dir```. For this example, I choosed ```tmp``` directory from my machine.
Also, there are other flags that we can use:

```-memprof_interval``` - Time between memory profile saves. By default, the time between writing into files is 5 seconds.
```-memprof_nrfiles```  - Number of memory profile to write. By default, the numbers of files is 1.

::

   cgr-engine -config_path=. -logger=*stdout -memprof_dir=/tmp/

In the running process, this will create a file named ```mem1_prof.prof```. Let the engine run for some time, and then you can kill the process. When the process is killed, it will create another file named ```mem_final.prof``` containing the final memory profiling information.

Reading the memory profile files
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Next step is to read properly the memory profiling. We will use the tool from go package, pprof. Go to the directory where the files were written and read the files.

::

   cd /tmp
   go tool pprof mem1.prof

It will open a console, and to check the memory usage, type ```top``` to see. For more documentation about this tool, we recommand you to read the pprof package documentation from golang, or simply type ```help```.