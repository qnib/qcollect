/*
Package collector contains the actual qcollect collectors (and the corresponding tests). All collectors need to embed baseCollector. Look at one of the existing collectors (test.go) to see how this is done.

mesos collector (mesos.go): This collector runs on all mesos masters. It identifies the leader amongst masters and collects stats from this leader only. Mesos masters report stats on :5050/metrics/snapshot, which is JSON. All these stats are pushed via qcollect to the configured handlers. Some sanitization is performed to convert the names to a more metric-y style. For example, "masters/cpus" would be changed to "masters.cpu."
*/
package collector
