package config

import (
	"testing"

	mariadbv1alpha1 "github.com/mariadb-operator/mariadb-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigMarshal(t *testing.T) {
	tests := []struct {
		name       string
		mariadb    *mariadbv1alpha1.MariaDB
		env        map[string]string
		wantConfig string
		wantErr    bool
	}{
		{
			name: "no env",
			mariadb: &mariadbv1alpha1.MariaDB{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mariadb-galera",
					Namespace: "default",
				},
				Spec: mariadbv1alpha1.MariaDBSpec{
					Galera: &mariadbv1alpha1.Galera{
						SST:            mariadbv1alpha1.SSTRsync,
						ReplicaThreads: 1,
					},
					Replicas: 3,
				},
			},
			env:        map[string]string{},
			wantConfig: "",
			wantErr:    true,
		},
		{
			name: "rsync",
			mariadb: &mariadbv1alpha1.MariaDB{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mariadb-galera",
					Namespace: "default",
				},
				Spec: mariadbv1alpha1.MariaDBSpec{
					Galera: &mariadbv1alpha1.Galera{
						SST:            mariadbv1alpha1.SSTRsync,
						ReplicaThreads: 1,
					},
					Replicas: 3,
				},
			},
			env: map[string]string{
				"HOSTNAME":              "mariadb-galera-0",
				"MARIADB_ROOT_PASSWORD": "foo",
			},
			//nolint:lll
			wantConfig: `[mysqld]
bind-address=0.0.0.0
default_storage_engine=InnoDB
binlog_format=row
innodb_autoinc_lock_mode=2

# Cluster configuration
wsrep_on=ON
wsrep_provider=/usr/lib/galera/libgalera_smm.so
wsrep_cluster_address="gcomm://mariadb-galera-0.mariadb-galera-internal.default.svc.cluster.local,mariadb-galera-1.mariadb-galera-internal.default.svc.cluster.local,mariadb-galera-2.mariadb-galera-internal.default.svc.cluster.local"
wsrep_cluster_name=mariadb-operator
wsrep_slave_threads=1

# Node configuration
wsrep_node_address="mariadb-galera-0.mariadb-galera-internal.default.svc.cluster.local"
wsrep_node_name="mariadb-galera-0"
wsrep_sst_method="rsync"
`,
			wantErr: false,
		},
		{
			name: "mariabackup",
			mariadb: &mariadbv1alpha1.MariaDB{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mariadb-galera",
					Namespace: "default",
				},
				Spec: mariadbv1alpha1.MariaDBSpec{
					Galera: &mariadbv1alpha1.Galera{
						SST:            mariadbv1alpha1.SSTMariaBackup,
						ReplicaThreads: 2,
					},
					Replicas: 3,
				},
			},
			env: map[string]string{
				"HOSTNAME":              "mariadb-galera-1",
				"MARIADB_ROOT_PASSWORD": "foo",
			},
			//nolint:lll
			wantConfig: `[mysqld]
bind-address=0.0.0.0
default_storage_engine=InnoDB
binlog_format=row
innodb_autoinc_lock_mode=2

# Cluster configuration
wsrep_on=ON
wsrep_provider=/usr/lib/galera/libgalera_smm.so
wsrep_cluster_address="gcomm://mariadb-galera-0.mariadb-galera-internal.default.svc.cluster.local,mariadb-galera-1.mariadb-galera-internal.default.svc.cluster.local,mariadb-galera-2.mariadb-galera-internal.default.svc.cluster.local"
wsrep_cluster_name=mariadb-operator
wsrep_slave_threads=2

# Node configuration
wsrep_node_address="mariadb-galera-1.mariadb-galera-internal.default.svc.cluster.local"
wsrep_node_name="mariadb-galera-1"
wsrep_sst_method="mariabackup"
wsrep_sst_auth="root:foo"
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfigFile(tt.mariadb)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			bytes, err := config.Marshal()
			if tt.wantErr && err == nil {
				t.Fatal("error expected, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("error unexpected, got %v", err)
			}
			if tt.wantConfig != string(bytes) {
				t.Fatalf("unexpected result:\nexpected:\n%s\ngot:\n%s\n", tt.wantConfig, string(bytes))
			}
		})
	}
}
