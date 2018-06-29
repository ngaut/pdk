// Copyright 2017 Pilosa Corp.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
// contributors may be used to endorse or promote products derived
// from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH
// DAMAGE.

package pdk_test

import (
	"testing"
	"time"

	gopilosa "github.com/pilosa/go-pilosa"
	"github.com/pilosa/pdk"
	ptest "github.com/pilosa/pilosa/test"
)

func TestSetupPilosa(t *testing.T) {
	s := ptest.MustRunCluster(t, 2)
	hosts := []string{}
	for _, com := range s {
		hosts = append(hosts, com.Server.URI.String())
	}

	schema := gopilosa.NewSchema()
	index, err := schema.Index("newindex")
	if err != nil {
		t.Fatal(err)
	}
	index.Field("field1", gopilosa.OptFieldSet(gopilosa.CacheTypeRanked, 17))
	index.Field("field2", gopilosa.OptFieldSet(gopilosa.CacheTypeLRU, 19))
	index.Field("field3", gopilosa.OptFieldInt(0, 20000))
	index.Field("fieldtime", gopilosa.OptFieldTime(gopilosa.TimeQuantumYearMonthDay))

	indexer, err := pdk.SetupPilosa(hosts, index.Name(), schema, 2)
	if err != nil {
		t.Fatalf("SetupPilosa: %v", err)
	}

	indexer.AddBit("field1", 0, 0)
	indexer.AddValue("field3", 0, 97)
	indexer.AddBitTimestamp("fieldtime", 0, 0, time.Date(2018, time.February, 22, 9, 0, 0, 0, time.UTC))
	indexer.AddBitTimestamp("fieldtime", 2, 0, time.Date(2018, time.February, 24, 9, 0, 0, 0, time.UTC))
	indexer.AddValue("field3", 0, 100)

	err = indexer.Close()
	if err != nil {
		t.Fatalf("closing indexer: %v", err)
	}

	client, err := gopilosa.NewClient(hosts)
	if err != nil {
		t.Fatalf("getting client: %v", err)
	}
	schema, err = client.Schema()
	if err != nil {
		t.Fatalf("getting schema: %v", err)
	}

	idxs := schema.Indexes()
	if len(idxs) != 1 {
		t.Fatalf("too many indexes: %v", idxs)
	}
	if idx, ok := idxs["newindex"]; !ok {
		t.Fatalf("index with wrong name: %v", idx)
	}

	if len(idxs["newindex"].Fields()) != 4 {
		t.Fatalf("wrong number of fields: %v", idxs["newindex"].Fields())
	}

	idx, err := schema.Index("newindex")
	if err != nil {
		t.Fatalf("getting index: %v", err)
	}
	fieldtime, err := idx.Field("fieldtime")
	if err != nil {
		t.Fatalf("getting field: %v", err)
	}
	resp, err := client.Query(fieldtime.Range(0, time.Date(2018, time.February, 21, 9, 0, 0, 0, time.UTC), time.Date(2018, time.February, 23, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("executing range query: %v", err)
	}
	bits := resp.Result().Row().Columns
	if len(bits) != 1 || bits[0] != 0 {
		t.Fatalf("unexpected bits from range query: %v", bits)
	}

	resp, err = client.Query(fieldtime.Range(0, time.Date(2018, time.February, 20, 9, 0, 0, 0, time.UTC), time.Date(2018, time.February, 21, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("executing range query: %v", err)
	}
	bits = resp.Result().Row().Columns
	if len(bits) != 0 {
		t.Fatalf("unexpected bits from empty range query: %v", bits)
	}

	resp, err = client.Query(fieldtime.Range(0, time.Date(2018, time.February, 20, 9, 0, 0, 0, time.UTC), time.Date(2018, time.February, 25, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("executing range query: %v", err)
	}
	bits = resp.Result().Row().Columns
	if len(bits) != 2 || bits[1] != 2 || bits[0] != 0 {
		t.Fatalf("unexpected bits from empty range query: %v", bits)
	}

	field3, err := idx.Field("field3")
	if err != nil {
		t.Fatalf("getting field: %v", err)
	}

	resp, err = client.Query(field3.Equals(100))
	if err != nil {
		t.Fatalf("executing range query: %v", err)
	}
	bits = resp.Result().Row().Columns
	if len(bits) != 1 || bits[0] != 0 {
		t.Fatalf("unexpected bits from range field query: %v", bits)
	}

}
