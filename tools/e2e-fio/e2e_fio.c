#include <stdio.h>
#include <signal.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include <errno.h>

/* struct for linked list of child processes */
typedef struct e2e_process {
    struct e2e_process* next;
    pid_t   pid;
    int     status;
    int     exitv;
    int     termsig;
    int     abnormal_exit;
    int     finished;
    char*   cmd;
} e2e_process;

/* head of the linked list of child processes */
static e2e_process* proc_list = NULL;

/* Create and run fio in child processes as defined in the list */
void start_procs() {
    e2e_process* proc = proc_list;
    while (proc != NULL) {
        proc->pid = fork();
        if ( 0 == proc->pid ) {
            /* Change working directory to avoid trivial collisions in file
             * space across multiple jobs
             */
            char wkspace[32];
            snprintf(wkspace, 32, "./%d", getpid());
            if (0 == mkdir(wkspace,0777) && (0 == chdir(wkspace)) ) {
                execl("/bin/sh", "sh", "-c", proc->cmd, NULL);
                printf("** execl %s failed %d **\n", proc->cmd, errno);
            } else {
               printf("** mkdir %s failed **\n", wkspace);
            }
            exit(errno);
        }
        printf("pid:%d, %s\n", proc->pid, proc->cmd);
        proc = proc->next;
    }
}

/*
 * parse command line arguments and populate a single e2e processes argument list,
 * and append it to the global list of e2e processes
 */
char** parse_procs(char **argv) {
    e2e_process *proc = NULL;
    /* Tis' C so we do it the "hard way" */
    char *pinsert;
    const char *executable = "fio ";
    size_t buflen = 0;

    /* 1. work out the size of the buffer required to copy the arguments.*/
    for(char **argv_scan=argv; *argv_scan != NULL; ++argv_scan) {
        /* +1 for space delimiter */
        buflen += strlen(*argv_scan) + 1;
    }

    if (buflen == 0) {
        return NULL;
    }

    proc = calloc(sizeof(*proc), 1);
    if (proc == NULL) {
        puts("failed to allocate memory for e2e_process");
        return NULL;
    }

    buflen += strlen(executable) + 1;
    /* 2. allocate a 0 intialised buffer so we can use strcat */
    proc->cmd = calloc(sizeof(unsigned char), buflen);
    if (proc->cmd == NULL) {
        free(proc);
        puts("failed to allocate memory for command line");
        return NULL;
    }

    pinsert = proc->cmd;
    /* 3. construct the command line, using strcat */
    strcat(pinsert, executable);
    pinsert += strlen(pinsert);
    for(; *argv != NULL && (0 != strcmp(*argv, "--")); ++argv) {
        strcat(pinsert, *argv);
        pinsert += strlen(pinsert);
        *pinsert = ' ';
        ++pinsert;
    }

    /* 4 append the process to the list */
    {
        e2e_process** insert_proc = &proc_list;
        while (*insert_proc != NULL) {
            insert_proc = &(*insert_proc)->next;
        }
        *insert_proc = proc;
    }

    return argv;
}

/* Kill all processes as defined in the list */
void kill_procs(int signal) {
    for (e2e_process* proc_ptr = proc_list; NULL != proc_ptr; proc_ptr = proc_ptr->next) {
        if (0 != proc_ptr->pid) {
            kill(proc_ptr->pid, signal);
        }
    }
}


/* Wait for all processes in the list to complete.*/
int wait_procs() {
    int exitv = 0;
    int pending;

    do {
        sleep(2);
        pending = 0;
        for (e2e_process* proc_ptr = proc_list; NULL != proc_ptr; proc_ptr = proc_ptr->next) {
            if (proc_ptr->finished) {
                continue;
            }

            if ( 0 == waitpid(proc_ptr->pid, &proc_ptr->status, WNOHANG)) {
                pending += 1;
                continue;
            }

            proc_ptr->finished = 1;
            printf("** %d finished **\n", proc_ptr->pid);
            if (WIFEXITED(proc_ptr->status)) {
                proc_ptr->exitv = WEXITSTATUS(proc_ptr->status);
                if (0 != proc_ptr->exitv) {
                    printf("** exit value = %d for %s **\n", proc_ptr->exitv, proc_ptr->cmd);
                }
            } else if (WIFSIGNALED(proc_ptr->status)) {
                proc_ptr->termsig = WTERMSIG(proc_ptr->status);
                printf("** termsig %d, %s **\n", proc_ptr->termsig, proc_ptr->cmd);
            } else {
                /* Should not reach here */
                printf("** Bug in handling waitpid status **\n");
                proc_ptr->abnormal_exit = 1;
            }
        }
    }while(pending);

    for (e2e_process* proc_ptr = proc_list; NULL != proc_ptr; proc_ptr = proc_ptr->next) {
        if (proc_ptr->exitv) {
            exitv = proc_ptr->exitv;
        } else if (proc_ptr->termsig) {
            exitv = 254;
        } else if (proc_ptr->abnormal_exit) {
            exitv = 255;
        }
    }
    return exitv;
}

/* Print contents of processes in the list. */
void print_procs() {
    for (e2e_process* proc_ptr = proc_list; NULL != proc_ptr; proc_ptr = proc_ptr->next) {
        printf("pid:%d, status=%d, exit=%d, termsig=%d, abnormal_exit=%d finished=%d\ncmd=%s\n",
                proc_ptr->pid,
                proc_ptr->status,
                proc_ptr->exitv,
                proc_ptr->termsig,
                proc_ptr->abnormal_exit,
                proc_ptr->finished,
                proc_ptr->cmd);
    }
}

/*
 * Usage:
 * [sleep <sleep seconds>] [segfault-after <delay seconds>] [-- <fio args1> -- <fio args2> ....] [exitv <exit value>]
 * 1. fio is only run if fio arguments are specified.
 * 2. fio is always run as a forked process.
 * 3. the segfault directive takes priority over the sleep directive
 * 4. exitv <v> override exit value - this is to simulate failure in the test pod.
 * 5. argument parsing is simple, invalid specifications are skipped over
 *  for example "sleep --" => sleep is skipped over, parsing resumes from "--"
 */
int main(int argc, char **argv_in)
{
    unsigned sleep_time = 0;
    unsigned segfault_time = 0;
    char** argv = argv_in;
    int     exitv = 0;
    int     procs_exitv = 0;

    puts("e2e_fio: version 3.00");
    for(char **argv_scan=argv; *argv_scan != NULL; ++argv_scan) {
        printf("%s ",*argv_scan);
    }
    puts("\n");
    fflush(stdout);

    /* skip over this programs name */
    argv += 1;
    /* "parse" arguments -
     * 1. segfault-after <n> number of seconds
     * 2. sleep <n> number of seconds
     * 3. anything after "--" is collected and passed to fio as arguments
     *    -- is also the separator for argument lists for multiple fio instances
     *
     * 4. exitv <v> override exit value - this is to aid test development.
     *      specifically to validate error detection in the tests.
     * For our simple purposes atoi is sufficient
     *
     * For simplicity none of the arguments are mandatory
     * if no arguments are supplied execution ends
     * segfault-after is always handled after sleep
     * fio is always run as a forked process executes concurrently
     * if fio instances are launched, we wait for all to complete and return 0 if
     * all instances ran succesfully, or the exit value of the last failing fio instance.
     * Note 1) you can use --status-interval=N as an argument to get fio to print status every N seconds
     *      2) output will be garbled when running multiple instances of fio.
     *
     * Intended use cases are
     * a) sleep N fio is executed using exec
     * b) segfault-after N, sleep for N then segfault terminating the pod or restarting the pod
     * c) segfault-after N -- ...., run fio (in a different process) and segfault after N seconds
     * d) -- ....., run fio, if fio completes, execution ends.
     */
    while(*argv != NULL) {
        if (0 == strcmp(*argv,"sleep") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            sleep_time = atoi(*(argv+1));
            ++argv;
        } else if (0 == strcmp(*argv,"segfault-after") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            segfault_time = atoi(*(argv+1));
            ++argv;
        } else if (0 == strcmp(*argv, "--")) {
            break;
        } else if (0 == strcmp(*argv,"exitv") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            exitv = atoi(*(argv+1));
            printf("Overriding exit value to %d\n", exitv);
            ++argv;
        } else {
            printf("Ignoring %s\n", *argv);
        }
        ++argv;
    }

    /* parse the argument list and populate the list of fio processes to run */
    while (*argv != NULL) {
        if (0 == strcmp(*argv, "--")) {
            argv++;
            argv = parse_procs(argv);
        } else {
            break;
        }
    }

    /* start the fio processes */
    start_procs();

    if (0 != segfault_time) {
        printf("Segfaulting after %d seconds\n", segfault_time);
        sleep(segfault_time);
        if (NULL != proc_list) {
            kill_procs(SIGKILL);
            sleep(1);
        }
        puts("Segfaulting now!");
        raise(SIGSEGV);
    }

    /* sleep takes priority over other actions,
     * in particular the use case is to allow,
     * a bunch of pods to startup
     * prior to loading the node by running fio
     */
    if (0 != sleep_time) {
        printf("sleeping %d seconds\n", sleep_time);
        sleep(sleep_time);
    }


    procs_exitv = wait_procs();
    puts("");
    print_procs();

    if (exitv == 0) {
        exitv = procs_exitv;
    }

    printf("Exit value is %d\n", exitv);
    return exitv;
}
