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
void start_proc(e2e_process* proc_ptr ) {
    proc_ptr->pid = fork();
    if ( 0 == proc_ptr->pid ) {
        /* Change working directory to avoid trivial collisions in file
         * space across multiple jobs
         */
        char wkspace[32];
        snprintf(wkspace, 32, "./%d", getpid());
        if (0 == mkdir(wkspace,0777) && (0 == chdir(wkspace)) ) {
            execl("/bin/sh", "sh", "-c", proc_ptr->cmd, NULL);
            printf("** execl %s failed %d **\n", proc_ptr->cmd, errno);
        } else {
            printf("** mkdir %s failed **\n", wkspace);
        }
        exit(errno);
    }
    printf("pid:%d, %s\n", proc_ptr->pid, proc_ptr->cmd);
    fflush(stdout);
}

/*
 * parse command line arguments and populate a single e2e processes argument list,
 * and append it to the global list of e2e processes
 */
char** parse_procs(char **argv, char command[]) {
    e2e_process *proc_ptr = NULL;
    /* Tis' C so we do it the "hard way" */
    char *pinsert;
    const char *executable = command;
    size_t buflen = 0;

    /* 1. work out the size of the buffer required to copy the arguments.*/
    for(char **argv_scan=argv; *argv_scan != NULL && (0 != strcmp(*argv_scan, "&")); ++argv_scan) {
        /* +1 for space delimiter */
        buflen += strlen(*argv_scan) + 1;
    }

    buflen += strlen(executable) + 1;
    if (buflen == 0) {
        return NULL;
    }
    proc_ptr = calloc(sizeof(*proc_ptr), 1);
    if (proc_ptr == NULL) {
        puts("failed to allocate memory for e2e_process");
        return NULL;
    }

    /* 2. allocate a 0 intialised buffer so we can use strcat */
    proc_ptr->cmd = calloc(sizeof(unsigned char), buflen);
    if (proc_ptr->cmd == NULL) {
        free(proc_ptr);
        puts("failed to allocate memory for command line");
        return NULL;
    }

    pinsert = proc_ptr->cmd;
    /* 3. construct the command line, using strcat */
    strcat(pinsert, executable);
    strcat(pinsert, " ");
    pinsert += strlen(pinsert);
    for(; *argv != NULL && (0 != strcmp(*argv, "&")); ++argv) {
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
        *insert_proc = proc_ptr;
    }

    start_proc(proc_ptr);
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
            fflush(stdout);
        }
    } while(pending);

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
 * [sleep <sleep seconds>] [segfault-after <delay seconds>] [-- <fio args1> & -- <fio args2> & ....] [exitv <exit value>]
 * 1. All arguments can be specified multiple times and will be executed in sequence.
 * 2. fio is only run if fio arguments are specified.
 * 3. fio is always run as a forked process.
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

    puts("e2e_fio: version 3.03");
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
     */
    while(*argv != NULL) {
        if (0 == strcmp(*argv,"sleep") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            sleep_time = atoi(*(argv+1));
            printf("sleeping %d seconds\n", sleep_time);
            sleep(sleep_time);
            ++argv;
        } else if (0 == strcmp(*argv,"segfault-after") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            segfault_time = atoi(*(argv+1));
            printf("Segfaulting after %d seconds\n", segfault_time);
            sleep(segfault_time);
            if (NULL != proc_list) {
                kill_procs(SIGKILL);
                sleep(1);
            }
            puts("Segfaulting now!");
            raise(SIGSEGV);
            ++argv;
        } else if (0 == strcmp(*argv, "--")) {
            char **next = parse_procs(argv+1,"fio");
            if (*next == NULL) {
                argv = next - 1;
            } else {
                argv = next;
            }

        } else if (0 == strcmp(*argv, "---")) {
            char **next = parse_procs(argv+2,*(argv+1));
            if (*next == NULL) {
                argv = next - 1;
            } else {
                argv = next;
            }

        } else if (0 == strcmp(*argv, "command")) {
            argv++;
            size_t buf = 0;
            for(char **argv_scan=argv; *argv_scan != NULL && (0 != strcmp(*argv_scan, "+")); ++argv_scan) {
                buf += strlen(*argv_scan) + 1;
            }
            char *cmdline;
            char *pinsert;
            cmdline = calloc(sizeof(unsigned char), buf);
            if (cmdline != NULL) {
                pinsert = cmdline;
            }else{
                printf("ERROR: failed to allocate memory\n");
                exit(-1);
            }
            for(; *argv != NULL && (0 != strcmp(*argv, "+")); ++argv) {
                strcat(pinsert, *argv);
                pinsert += strlen(pinsert);
                *pinsert = ' ';
                ++pinsert;
            }
            if (system(cmdline)!= 0) {
                printf("ERROR: system call failed with %d\n%s",errno, cmdline);
                exit(-1);
            }
            free(cmdline);
        } else if (0 == strcmp(*argv,"exitv") && NULL != *(argv+1) && 0 != atoi(*(argv+1))) {
            exitv = atoi(*(argv+1));
            printf("Overriding exit value to %d\n", exitv);
            ++argv;
        } else {
            printf("Ignoring %s\n", *argv);
        }
        ++argv;
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
