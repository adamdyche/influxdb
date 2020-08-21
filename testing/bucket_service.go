package testing

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/mock"
)

const (
	idOne = influxdb.ID(iota + 1)
	idTwo
	idThree
	idFour
	idFive
)

var bucketCmpOptions = cmp.Options{
	cmp.Comparer(func(x, y []byte) bool {
		return bytes.Equal(x, y)
	}),
	cmp.Comparer(func(x, y *influxdb.Bucket) bool {
		if x == nil && y == nil {
			return true
		}
		if x != nil && y == nil || y != nil && x == nil {
			return false
		}

		return x.OrgID == y.OrgID &&
			x.Type == y.Type &&
			x.Description == y.Description &&
			x.RetentionPolicyName == y.RetentionPolicyName &&
			x.RetentionPeriod == y.RetentionPeriod &&
			x.Name == y.Name
	}),
	cmp.Transformer("Sort", func(in []*influxdb.Bucket) []*influxdb.Bucket {
		out := append([]*influxdb.Bucket(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			return out[i].Name > out[j].Name
		})
		return out
	}),
}

var systemBucketCmpOptions = cmp.Options{
	cmpopts.IgnoreFields(influxdb.Bucket{}, "ID", "CreatedAt", "UpdatedAt"),
	cmp.Comparer(func(x, y []byte) bool {
		return bytes.Equal(x, y)
	}),
	cmp.Comparer(func(x, y *influxdb.Bucket) bool {
		if x == nil && y == nil {
			return true
		}
		if x != nil && y == nil || y != nil && x == nil {
			return false
		}

		return x.OrgID == y.OrgID &&
			x.Type == y.Type &&
			x.Description == y.Description &&
			x.RetentionPolicyName == y.RetentionPolicyName &&
			x.RetentionPeriod == y.RetentionPeriod &&
			x.Name == y.Name
	}),
}

// BucketFields will include the IDGenerator, and buckets
type BucketFields struct {
	IDGenerator   influxdb.IDGenerator
	OrgIDs        influxdb.IDGenerator
	BucketIDs     influxdb.IDGenerator
	TimeGenerator influxdb.TimeGenerator
	Buckets       []*influxdb.Bucket
	Organizations []*influxdb.Organization
}

type bucketServiceF func(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
)

// BucketService tests all the service functions.
func BucketService(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T) {
	tests := []struct {
		name string
		fn   bucketServiceF
	}{
		// {
		// 	name: "CreateBucket",
		// 	fn:   CreateBucket,
		// },
		// {
		// 	name: "FindBucketByID",
		// 	fn:   FindBucketByID,
		// },
		// {
		// 	name: "FindBuckets",
		// 	fn:   FindBuckets,
		// },
		{
			name: "FindSystemBuckets",
			fn:   FindSystemBuckets,
		},
		// {
		// 	name: "FindBucket",
		// 	fn:   FindBucket,
		// },
		// {
		// 	name: "UpdateBucket",
		// 	fn:   UpdateBucket,
		// },
		// {
		// 	name: "DeleteBucket",
		// 	fn:   DeleteBucket,
		// },
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fn(init, t)
		})
	}
}

// CreateBucket testing
func CreateBucket(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		bucket *influxdb.Bucket
	}
	type wants struct {
		err     error
		buckets []*influxdb.Bucket
	}

	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "create buckets with empty set",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idOne),
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Buckets:       []*influxdb.Bucket{},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:        "name1",
					OrgID:       idOne,
					Description: "desc1",
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						Name:        "name1",
						ID:          idOne,
						OrgID:       idOne,
						Description: "desc1",
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
							UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
		{
			name: "basic create bucket",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idTwo),
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "bucket1",
						OrgID: idOne,
					},
				},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:  "bucket2",
					OrgID: idTwo,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						Name:  "bucket1",
						OrgID: idOne,
					},
					{
						ID:    idTwo,
						Name:  "bucket2",
						OrgID: idTwo,
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
							UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
		{
			name: "names should be unique within an organization",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idTwo),
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "bucket1",
						OrgID: idOne,
					},
				},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:  "bucket1",
					OrgID: idOne,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						Name:  "bucket1",
						OrgID: idOne,
					},
				},
				err: &influxdb.Error{
					Code: influxdb.EConflict,
					Op:   influxdb.OpCreateBucket,
					Msg:  "bucket with name bucket1 already exists",
				},
			},
		},
		{
			name: "names should not be unique across organizations",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idTwo),
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "bucket1",
						OrgID: idOne,
					},
				},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:  "bucket1",
					OrgID: idTwo,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						Name:  "bucket1",
						OrgID: idOne,
						// CRUDLog is missing because seed data is created through
						// storage layer and not service layer (where CRUDLog is populated)
					},
					{
						ID:    idTwo,
						Name:  "bucket1",
						OrgID: idTwo,
						CRUDLog: influxdb.CRUDLog{
							CreatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
							UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
						},
					},
				},
			},
		},
		{
			name: "create bucket with orgID not exist",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idOne),
				OrgIDs:        mock.NewStaticIDGenerator(idOne),
				BucketIDs:     mock.NewStaticIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Buckets:       []*influxdb.Bucket{},
				Organizations: []*influxdb.Organization{},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:  "name1",
					OrgID: idOne,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{},
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Msg:  "organization not found",
					Op:   influxdb.OpCreateBucket,
				},
			},
		},
		{
			name: "create bucket with illegal quotation mark",
			fields: BucketFields{
				IDGenerator:   mock.NewStaticIDGenerator(idOne),
				BucketIDs:     mock.NewStaticIDGenerator(idOne),
				OrgIDs:        mock.NewStaticIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Buckets:       []*influxdb.Bucket{},
				Organizations: []*influxdb.Organization{
					{
						Name: "org",
					},
				},
			},
			args: args{
				bucket: &influxdb.Bucket{
					Name:  "namewith\"quote",
					OrgID: idOne,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{},
				err: &influxdb.Error{
					Code: influxdb.EInvalid,
					Op:   influxdb.OpCreateBucket,
					Msg:  "bucket name namewith\"quote is invalid. Bucket names may not include quotation marks",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			err := s.CreateBucket(ctx, tt.args.bucket)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			// Delete only newly created buckets - ie., with a not nil ID
			// if tt.args.bucket.ID.Valid() {
			defer s.DeleteBucket(ctx, tt.args.bucket.ID)
			// }

			buckets, _, err := s.FindBuckets(ctx, influxdb.BucketFilter{})
			if err != nil {
				t.Fatalf("failed to retrieve buckets: %v", err)
			}

			// remove system buckets
			filteredBuckets := []*influxdb.Bucket{}
			for _, b := range buckets {
				if b.Type != influxdb.BucketTypeSystem {
					filteredBuckets = append(filteredBuckets, b)
				}
			}

			if diff := cmp.Diff(filteredBuckets, tt.wants.buckets, bucketCmpOptions...); diff != "" {
				t.Errorf("buckets are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindBucketByID testing
func FindBucketByID(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		id influxdb.ID
	}
	type wants struct {
		err    error
		bucket *influxdb.Bucket
	}

	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "basic find bucket by id",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
						ID:   idOne,
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id: idOne,
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:    idOne,
					OrgID: idOne,
					Name:  "bucket1",
				},
			},
		},
		{
			name: "find bucket by id not exist",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id: idThree,
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Op:   influxdb.OpFindBucketByID,
					Msg:  "bucket not found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			bucket, err := s.FindBucketByID(ctx, tt.args.id)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			if diff := cmp.Diff(bucket, tt.wants.bucket, bucketCmpOptions...); diff != "" {
				t.Errorf("bucket is different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindBuckets testing
func FindBuckets(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		ID             influxdb.ID
		name           string
		organization   string
		organizationID influxdb.ID
		findOptions    influxdb.FindOptions
	}

	type wants struct {
		buckets []*influxdb.Bucket
		err     error
	}
	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "find all buckets",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						OrgID: idOne,
						Name:  "abc",
					},
					{
						OrgID: idTwo,
						Name:  "xyz",
					},
				},
			},
			args: args{},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						OrgID: idOne,
						Name:  "abc",
					},
					{
						ID:    idTwo,
						OrgID: idTwo,
						Name:  "xyz",
					},
				},
			},
		},
		{
			name: "find all buckets by offset and limit",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "def",
					},
					{
						// ID(3)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				findOptions: influxdb.FindOptions{
					Offset: 1,
					Limit:  1,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idTwo,
						OrgID: idOne,
						Name:  "def",
					},
				},
			},
		},
		{
			name: "find all buckets by after and limit",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "def",
					},
					{
						// ID(3)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				findOptions: influxdb.FindOptions{
					After: idPtr(idOne),
					Limit: 2,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idTwo,
						OrgID: idOne,
						Name:  "def",
					},
					{
						ID:    idThree,
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
		},
		{
			name: "find all buckets by descending",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "def",
					},
					{
						// ID(3)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				findOptions: influxdb.FindOptions{
					Offset:     1,
					Descending: true,
				},
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idTwo,
						OrgID: idOne,
						Name:  "def",
					},
					{
						ID:    idOne,
						OrgID: idOne,
						Name:  "abc",
					},
				},
			},
		},
		{
			name: "find buckets by organization name",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idTwo,
						Name:  "xyz",
					},
					{
						// ID(3)
						OrgID: idOne,
						Name:  "123",
					},
				},
			},
			args: args{
				organization: "theorg",
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						OrgID: idOne,
						Name:  "abc",
					},
					{
						ID:    idThree,
						OrgID: idOne,
						Name:  "123",
					},
				},
			},
		},
		{
			name: "find buckets by organization id",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
					{
						// ID(2)
						Name: "otherorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idTwo,
						Name:  "xyz",
					},
					{
						// ID(3)
						OrgID: idOne,
						Name:  "123",
					},
				},
			},
			args: args{
				organizationID: idOne,
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idOne,
						OrgID: idOne,
						Name:  "abc",
					},
					{
						ID:    idThree,
						OrgID: idOne,
						Name:  "123",
					},
				},
			},
		},
		{
			name: "find bucket by name",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				name: "xyz",
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						ID:    idTwo,
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
		},
		{
			name: "missing bucket returns no buckets",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{},
			},
			args: args{
				name: "xyz",
			},
			wants: wants{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			filter := influxdb.BucketFilter{}
			if tt.args.ID.Valid() {
				filter.ID = &tt.args.ID
			}
			if tt.args.organizationID.Valid() {
				filter.OrganizationID = &tt.args.organizationID
			}
			if tt.args.organization != "" {
				filter.Org = &tt.args.organization
			}
			if tt.args.name != "" {
				filter.Name = &tt.args.name
			}

			buckets, _, err := s.FindBuckets(ctx, filter, tt.args.findOptions)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			// remove system buckets
			filteredBuckets := []*influxdb.Bucket{}
			for _, b := range buckets {
				if b.Type != influxdb.BucketTypeSystem {
					filteredBuckets = append(filteredBuckets, b)
				}
			}

			if diff := cmp.Diff(filteredBuckets, tt.wants.buckets, bucketCmpOptions...); diff != "" {
				t.Errorf("buckets are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindSystemBuckets testing
func FindSystemBuckets(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		ID             influxdb.ID
		name           string
		organization   string
		organizationID influxdb.ID
		findOptions    influxdb.FindOptions
	}

	type wants struct {
		buckets []*influxdb.Bucket
		err     error
	}
	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		// {
		// 	name: "find legacy task and monitoring buckets",
		// 	fields: BucketFields{
		// 		OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
		// 		BucketIDs: mock.NewIncrementingIDGenerator(idOne),
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				// ID(1)
		// 				Name: "theorg",
		// 			},
		// 		},
		// 		Buckets: []*influxdb.Bucket{
		// 			{
		// 				OrgID: idOne,
		// 				ID:    idTwo,
		// 				Name:  "abc",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		organizationID: idOne,
		// 	},
		// 	wants: wants{
		// 		buckets: []*influxdb.Bucket{
		// 			{
		// 				Name:            influxdb.TasksSystemBucketName,
		// 				ID:              influxdb.TasksSystemBucketID,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.TasksSystemBucketRetention,
		// 				Description:     "System bucket for task logs",
		// 			},
		// 			{
		// 				Name:            influxdb.MonitoringSystemBucketName,
		// 				ID:              influxdb.MonitoringSystemBucketID,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.MonitoringSystemBucketRetention,
		// 				Description:     "System bucket for monitoring logs",
		// 			},
		// 			{
		// 				ID:    idTwo,
		// 				OrgID: idOne,
		// 				Name:  "abc",
		// 			},
		// 		},
		// 	},
		// },
		{
			name: "offset and limit automatically adjusted when adding legacy buckets",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						OrgID: idOne,
						ID:    idTwo,
						Name:  "aaa",
					},
					{
						OrgID: idOne,
						ID:    idThree,
						Name:  "abc",
					},
					{
						OrgID: idOne,
						ID:    idFour,
						Name:  "xyz",
					},
					{
						OrgID: idOne,
						ID:    idFive,
						Name:  "zzz",
					},
				},
			},
			args: args{
				organizationID: idOne,
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						Name:            influxdb.TasksSystemBucketName,
						ID:              influxdb.TasksSystemBucketID,
						Type:            influxdb.BucketTypeSystem,
						RetentionPeriod: influxdb.TasksSystemBucketRetention,
						Description:     "System bucket for task logs",
					},
					{
						Name:            influxdb.MonitoringSystemBucketName,
						ID:              influxdb.MonitoringSystemBucketID,
						Type:            influxdb.BucketTypeSystem,
						RetentionPeriod: influxdb.MonitoringSystemBucketRetention,
						Description:     "System bucket for monitoring logs",
					},
					{
						OrgID: idOne,
						ID:    idTwo,
						Name:  "aaa",
					},
					{
						OrgID: idOne,
						ID:    idThree,
						Name:  "abc",
					},
				},
			},
		},
		// {
		// 	name: "do not create duplicate legacy task and monitoring buckets",
		// 	fields: BucketFields{
		// 		OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
		// 		BucketIDs: mock.NewIncrementingIDGenerator(idOne),
		// 		Organizations: []*influxdb.Organization{
		// 			{
		// 				// ID(1)
		// 				Name: "theorg",
		// 			},
		// 		},
		// 		Buckets: []*influxdb.Bucket{
		// 			{
		// 				Name:            influxdb.MonitoringSystemBucketName,
		// 				OrgID:           idOne,
		// 				ID:              influxdb.MonitoringSystemBucketID,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.MonitoringSystemBucketRetention,
		// 				Description:     "System bucket for monitoring logs",
		// 			},
		// 			{
		// 				Name:            influxdb.TasksSystemBucketName,
		// 				OrgID:           idOne,
		// 				ID:              influxdb.TasksSystemBucketID,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.TasksSystemBucketRetention,
		// 				Description:     "System bucket for task logs",
		// 			},
		// 			{
		// 				OrgID: idOne,
		// 				ID:    idTwo,
		// 				Name:  "abc",
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		organizationID: idOne,
		// 	},
		// 	wants: wants{
		// 		buckets: []*influxdb.Bucket{
		// 			{
		// 				Name:            influxdb.MonitoringSystemBucketName,
		// 				ID:              influxdb.MonitoringSystemBucketID,
		// 				OrgID:           idOne,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.MonitoringSystemBucketRetention,
		// 				Description:     "System bucket for monitoring logs",
		// 			},
		// 			{
		// 				Name:            influxdb.TasksSystemBucketName,
		// 				ID:              influxdb.TasksSystemBucketID,
		// 				OrgID:           idOne,
		// 				Type:            influxdb.BucketTypeSystem,
		// 				RetentionPeriod: influxdb.TasksSystemBucketRetention,
		// 				Description:     "System bucket for task logs",
		// 			},
		// 			{
		// 				ID:    idTwo,
		// 				OrgID: idOne,
		// 				Name:  "abc",
		// 			},
		// 		},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			filter := influxdb.BucketFilter{}
			if tt.args.ID.Valid() {
				filter.ID = &tt.args.ID
			}
			if tt.args.organizationID.Valid() {
				filter.OrganizationID = &tt.args.organizationID
			}
			if tt.args.organization != "" {
				filter.Org = &tt.args.organization
			}
			if tt.args.name != "" {
				filter.Name = &tt.args.name
			}

			buckets, _, err := s.FindBuckets(ctx, filter, tt.args.findOptions)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			expected := tt.wants.buckets
			if n := len(buckets); n != len(expected) {
				t.Errorf("expected %d buckets, got: %d", len(expected), n)
			}

			for i, b := range expected {
				// fmt.Println("got order: ")
				// fmt.Println("got: ", buckets[i])
				// fmt.Println("wanted : ", b)
				if diff := cmp.Diff(buckets[i], b, systemBucketCmpOptions...); diff != "" {
					t.Errorf("buckets are different -got/+want\ndiff %s", diff)
				}
			}
		})
	}
}

// DeleteBucket testing
func DeleteBucket(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		ID influxdb.ID
	}
	type wants struct {
		err     error
		buckets []*influxdb.Bucket
	}

	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "delete buckets using exist id",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "A",
						OrgID: idOne,
					},
					{
						// ID(2)
						Name:  "B",
						OrgID: idOne,
					},
				},
			},
			args: args{
				ID: idOne,
			},
			wants: wants{
				buckets: []*influxdb.Bucket{
					{
						Name:  "B",
						ID:    idTwo,
						OrgID: idOne,
					},
				},
			},
		},
		{
			name: "delete buckets using id that does not exist",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "A",
						OrgID: idOne,
					},
					{
						// ID(2)
						Name:  "B",
						OrgID: idOne,
					},
				},
			},
			args: args{
				ID: MustIDBase16("1234567890654321"),
			},
			wants: wants{
				err: &influxdb.Error{
					Op:   influxdb.OpDeleteBucket,
					Msg:  "bucket not found",
					Code: influxdb.ENotFound,
				},
				buckets: []*influxdb.Bucket{
					{
						Name:  "A",
						ID:    idOne,
						OrgID: idOne,
					},
					{
						Name:  "B",
						ID:    idTwo,
						OrgID: idOne,
					},
				},
			},
		},
		{
			name: "delete system buckets",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "A",
						OrgID: idOne,
						Type:  influxdb.BucketTypeSystem,
					},
				},
			},
			args: args{
				ID: idOne,
			},
			wants: wants{
				err: &influxdb.Error{
					Op:   influxdb.OpDeleteBucket,
					Msg:  "system buckets cannot be deleted",
					Code: influxdb.EInvalid,
				},
				buckets: []*influxdb.Bucket{
					{
						Name:  "A",
						ID:    idOne,
						OrgID: idOne,
						Type:  influxdb.BucketTypeSystem,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			err := s.DeleteBucket(ctx, tt.args.ID)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			filter := influxdb.BucketFilter{}
			buckets, _, err := s.FindBuckets(ctx, filter)
			if err != nil {
				t.Fatalf("failed to retrieve buckets: %v", err)
			}

			// remove built in system buckets
			filteredBuckets := []*influxdb.Bucket{}
			for _, b := range buckets {
				if b.Name != influxdb.TasksSystemBucketName && b.Name != influxdb.MonitoringSystemBucketName {
					filteredBuckets = append(filteredBuckets, b)
				}
			}

			if diff := cmp.Diff(filteredBuckets, tt.wants.buckets, bucketCmpOptions...); diff != "" {
				t.Errorf("buckets are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// FindBucket testing
func FindBucket(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		name           string
		organizationID influxdb.ID
		id             influxdb.ID
	}

	type wants struct {
		bucket *influxdb.Bucket
		err    error
	}

	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "find bucket by name",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				name:           "abc",
				organizationID: idOne,
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:    idOne,
					OrgID: idOne,
					Name:  "abc",
				},
			},
		},
		{
			name: "find bucket by id",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "abc",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "xyz",
					},
				},
			},
			args: args{
				id:             idTwo,
				organizationID: idOne,
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:    idTwo,
					OrgID: idOne,
					Name:  "xyz",
				},
			},
		},
		{
			name: "missing bucket returns error",
			fields: BucketFields{
				OrgIDs:    mock.NewIncrementingIDGenerator(idOne),
				BucketIDs: mock.NewIncrementingIDGenerator(idOne),
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{},
			},
			args: args{
				name:           "xyz",
				organizationID: idOne,
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.ENotFound,
					Op:   influxdb.OpFindBucket,
					Msg:  "bucket \"xyz\" not found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()
			filter := influxdb.BucketFilter{}
			if tt.args.name != "" {
				filter.Name = &tt.args.name
			}
			if tt.args.id.Valid() {
				filter.ID = &tt.args.id
			}
			if tt.args.organizationID.Valid() {
				filter.OrganizationID = &tt.args.organizationID
			}

			bucket, err := s.FindBucket(ctx, filter)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			if diff := cmp.Diff(bucket, tt.wants.bucket, bucketCmpOptions...); diff != "" {
				t.Errorf("buckets are different -got/+want\ndiff %s", diff)
			}
		})
	}
}

// UpdateBucket testing
func UpdateBucket(
	init func(BucketFields, *testing.T) (influxdb.BucketService, string, func()),
	t *testing.T,
) {
	type args struct {
		name        string
		id          influxdb.ID
		retention   int
		description *string
	}
	type wants struct {
		err    error
		bucket *influxdb.Bucket
	}

	tests := []struct {
		name   string
		fields BucketFields
		args   args
		wants  wants
	}{
		{
			name: "update name",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:   idTwo,
				name: "changed",
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:    idTwo,
					OrgID: idOne,
					Name:  "changed",
					CRUDLog: influxdb.CRUDLog{
						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "update name unique",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:   idOne,
				name: "bucket2",
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.EConflict,
					Msg:  "bucket name is not unique",
				},
			},
		},
		{
			name: "update system bucket name",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Type:  influxdb.BucketTypeSystem,
						Name:  "bucket1",
					},
				},
			},
			args: args{
				id:   idOne,
				name: "bucket2",
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.EInvalid,
					Msg:  "system buckets cannot be renamed",
				},
			},
		},
		{
			name: "update retention",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:        idOne,
				retention: 100,
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:              idOne,
					OrgID:           idOne,
					Name:            "bucket1",
					RetentionPeriod: 100 * time.Minute,
					CRUDLog: influxdb.CRUDLog{
						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "update description",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:          idOne,
				description: stringPtr("desc1"),
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:          idOne,
					OrgID:       idOne,
					Name:        "bucket1",
					Description: "desc1",
					CRUDLog: influxdb.CRUDLog{
						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "update retention and name",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:        idOne,
				retention: 101,
				name:      "changed",
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:              idOne,
					OrgID:           idOne,
					Name:            "changed",
					RetentionPeriod: 101 * time.Minute,
					CRUDLog: influxdb.CRUDLog{
						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "update retention and same name",
			fields: BucketFields{
				OrgIDs:        mock.NewIncrementingIDGenerator(idOne),
				BucketIDs:     mock.NewIncrementingIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "theorg",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						OrgID: idOne,
						Name:  "bucket1",
					},
					{
						// ID(2)
						OrgID: idOne,
						Name:  "bucket2",
					},
				},
			},
			args: args{
				id:        idTwo,
				retention: 101,
				name:      "bucket2",
			},
			wants: wants{
				bucket: &influxdb.Bucket{
					ID:              idTwo,
					OrgID:           idOne,
					Name:            "bucket2",
					RetentionPeriod: 101 * time.Minute,
					CRUDLog: influxdb.CRUDLog{
						UpdatedAt: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "update bucket with illegal quotation mark",
			fields: BucketFields{
				OrgIDs:        mock.NewStaticIDGenerator(idOne),
				BucketIDs:     mock.NewStaticIDGenerator(idOne),
				TimeGenerator: mock.TimeGenerator{FakeValue: time.Date(2006, 5, 4, 1, 2, 3, 0, time.UTC)},
				Organizations: []*influxdb.Organization{
					{
						// ID(1)
						Name: "org",
					},
				},
				Buckets: []*influxdb.Bucket{
					{
						// ID(1)
						Name:  "valid name",
						OrgID: idOne,
					},
				},
			},
			args: args{
				id:   idOne,
				name: "namewith\"quote",
			},
			wants: wants{
				err: &influxdb.Error{
					Code: influxdb.EInvalid,
					Op:   influxdb.OpCreateBucket,
					Msg:  "bucket name namewith\"quote is invalid. Bucket names may not include quotation marks",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, opPrefix, done := init(tt.fields, t)
			defer done()
			ctx := context.Background()

			upd := influxdb.BucketUpdate{}
			if tt.args.name != "" {
				upd.Name = &tt.args.name
			}
			if tt.args.retention != 0 {
				d := time.Duration(tt.args.retention) * time.Minute
				upd.RetentionPeriod = &d
			}

			upd.Description = tt.args.description

			bucket, err := s.UpdateBucket(ctx, tt.args.id, upd)
			diffPlatformErrors(tt.name, err, tt.wants.err, opPrefix, t)

			if diff := cmp.Diff(bucket, tt.wants.bucket, bucketCmpOptions...); diff != "" {
				t.Errorf("bucket is different -got/+want\ndiff %s", diff)
			}
		})
	}
}
