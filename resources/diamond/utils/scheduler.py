# coding=utf-8

import time
import math
import multiprocessing
import os
import random
import sys
import signal
from subprocess import Popen, PIPE

try:
    import psutil
except ImportError:
    psutil = None

try:
    from setproctitle import getproctitle, setproctitle
except ImportError:
    setproctitle = None

from diamond.utils.signals import signal_to_exception
from diamond.utils.signals import SIGALRMException
from diamond.utils.signals import SIGHUPException


def get_children(parent_pid):
    if psutil:
        parent = psutil.Process(int(parent_pid))
        return [child.pid for child in parent.get_children()]
    else:
        children = []
        process = Popen(['ps', '-eo', 'pid,ppid'], stdout=PIPE, stderr=PIPE)
        output, errors = process.communicate()
        if errors:
            log.error("Could not get processlist with child procs: {0!s}".format(errors))
            return children
        for line in output.splitlines():
            pid, ppid = line.split(' ', 1)
            if ppid == parent_pid:
                children.append(pid)
        return children


def collector_process(collector, log):
    """
    """
    proc = multiprocessing.current_process()
    pid = str(proc.pid)
    if setproctitle:
        setproctitle('%s - %s' % (getproctitle(), proc.name))

    signal.signal(signal.SIGALRM, signal_to_exception)
    signal.signal(signal.SIGHUP, signal_to_exception)
    signal.signal(signal.SIGUSR2, signal_to_exception)

    interval = float(collector.config['interval'])

    log.debug('Starting')
    log.debug('Interval: %s seconds', interval)

    # Validate the interval
    if interval <= 0:
        log.critical('interval of %s is not valid!', interval)
        sys.exit(1)

    # Start the next execution at the next window plus some stagger delay to
    # avoid having all collectors running at the same time
    next_window = math.floor(time.time() / interval) * interval
    stagger_offset = random.uniform(0, interval - 1)

    # Allocate time till the end of the window for the collector to run. With a
    # minimum of 1 second
    max_time = int(max(interval - stagger_offset, 1))
    log.debug('Max collection time: %s seconds', max_time)

    # Setup stderr/stdout as /dev/null so random print statements in thrid
    # party libs do not fail and prevent collectors from running.
    # https://github.com/BrightcoveOS/Diamond/issues/722
    sys.stdout = open(os.devnull, 'w')
    sys.stderr = open(os.devnull, 'w')

    while(True):
        try:
            time_to_sleep = (next_window + stagger_offset) - time.time()
            if time_to_sleep > 0:
                time.sleep(time_to_sleep)

            next_window += interval

            # Ensure collector run times fit into the collection window
            signal.alarm(max_time)

            # Collect!
            collector._run()

            # Success! Disable the alarm
            signal.alarm(0)

        except SIGALRMException:

            # Adjust  the stagger_offset to allow for more time to run the
            # collector
            stagger_offset = stagger_offset * 0.9

            max_time = int(max(interval - stagger_offset, 1))
            log.debug('Max collection time: %s seconds', max_time)
            collector.dimensions = {
                'interval': interval,
            }
            collector.publish('fullerite.collection_time_exceeded', 1)
            try:
                collector.log.warn("Took too long to run, Killed!")
                children = get_children(pid)
                for child in children:
                    os.kill(int(child), signal.SIGKILL)
            except OSError as e:
                log.debug('Process died on its own!')
            except Exception as e:
                collector.log.warn("Killing children failed")

        except SIGHUPException:
            # Reload the config if requested
            # We must first disable the alarm as we don't want it to interrupt
            # us and end up with half a loaded config
            signal.alarm(0)

            log.info('Reloading config reload due to HUP')
            collector.load_config()
            log.info('Config reloaded')

        except Exception:
            log.exception('Collector failed!')
            break
