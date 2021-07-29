package apiclient

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/helm/pkg/urlutil"
	"strconv"
	"time"
)

type QuotaType string
type QuotaStatus string

const QuotaTypeHard QuotaType = "HARD"
const QuotaTypeSoft QuotaType = "SOFT"
const QuotaTypeDefault = QuotaTypeHard
const QuotaStatusActive = "ACTIVE"
const QuotaStatusPending = "ADDING"
const QuotaStatusError = "ERROR"
const QuotaStatusDeleting = "DELETING"
const MaxQuotaSize uint64 = 18446744073709547520

type Quota struct {
	FilesystemUid  uuid.UUID `json:"-"`
	InodeId        uint64    `json:"inodeId,omitempty"`
	TotalBytes     uint64    `json:"totalBytes,omitempty"`
	HardLimitBytes uint64    `json:"hardLimitBytes,omitempty"`
	SoftLimitBytes uint64    `json:"softLimitBytes,omitempty"`
	Status         string    `json:"status,omitempty"`
}

func (q *Quota) String() string {
	return fmt.Sprintln("Quota(fsUid:", q.FilesystemUid, "inodeId:", q.InodeId, "type:", q.GetQuotaType(), "capacity:", q.GetCapacityLimit(), "status:", q.Status, ")")
}

func (q *Quota) GetType() string {
	return "quota"
}

func (q *Quota) GetBasePath() string {
	fsUrl := (&FileSystem{Uid: q.FilesystemUid}).GetApiUrl()
	url, err := urlutil.URLJoin(fsUrl, q.GetType())
	if err != nil {
		return ""
	}
	return url
}

func (q *Quota) GetApiUrl() string {
	url, err := urlutil.URLJoin(q.GetBasePath(), strconv.FormatUint(q.InodeId, 10))
	if err != nil {
		return ""
	}
	return url
}

func (q *Quota) getImmutableFields() []string {
	return []string{
		"filesystemUid",
		"InodeId",
	}
}

func (q *Quota) EQ(r ApiObject) bool {
	return ObjectsAreEqual(r, q)
}

func (q *Quota) GetQuotaType() QuotaType {
	if q.HardLimitBytes < q.SoftLimitBytes {
		return QuotaTypeHard
	}
	return QuotaTypeSoft
}

func (q *Quota) GetCapacityLimit() uint64 {
	if q.GetQuotaType() == QuotaTypeHard {
		return q.HardLimitBytes
	}
	return q.SoftLimitBytes
}

type QuotaCreateRequest struct {
	filesystemUid  uuid.UUID
	InodeId        uint64 `json:"inodeId,omitempty"`
	HardLimitBytes uint64 `json:"hard_quota,omitempty"`
	SoftLimitBytes uint64 `json:"soft_quota,omitempty"`
	quotaType      QuotaType
	capacityLimit  uint64
}

func (qc *QuotaCreateRequest) getApiUrl() string {
	return qc.getRelatedObject().GetApiUrl()
}

func (qc *QuotaCreateRequest) getRequiredFields() []string {
	return []string{"InodeId", "filesystemUid", "quotaType", "capacityLimit"}
}
func (qc *QuotaCreateRequest) hasRequiredFields() bool {
	return ObjectRequestHasRequiredFields(qc)
}
func (qc *QuotaCreateRequest) getRelatedObject() ApiObject {
	return &Quota{
		FilesystemUid: qc.filesystemUid,
		InodeId:       qc.InodeId,
	}
}
func (qc *QuotaCreateRequest) String() string {
	return fmt.Sprintln("QuotaCreateRequest(fsUid:", qc.filesystemUid, "inodeId:", qc.InodeId, "type:", qc.quotaType, "capacity:", qc.capacityLimit, ")")
}

type QuotaUpdateRequest struct {
	filesystemUid  uuid.UUID
	InodeId        uint64 `json:"inodeId,omitempty"`
	HardLimitBytes uint64 `json:"hardLimitBytes,omitempty"`
	SoftLimitBytes uint64 `json:"softLimitBytes,omitempty"`
	quotaType      QuotaType
	capacityLimit  uint64
}

func (qu *QuotaUpdateRequest) getApiUrl() string {
	return qu.getRelatedObject().GetBasePath()
}

func (qu *QuotaUpdateRequest) getRequiredFields() []string {
	return []string{"InodeId", "filesystemUid", "quotaType", "capacityLimit"}
}
func (qu *QuotaUpdateRequest) hasRequiredFields() bool {
	return ObjectRequestHasRequiredFields(qu)
}
func (qu *QuotaUpdateRequest) getRelatedObject() ApiObject {
	return &Quota{
		FilesystemUid: qu.filesystemUid,
		InodeId:       qu.InodeId,
	}
}
func (qu *QuotaUpdateRequest) String() string {
	return fmt.Sprintln("QuotaUpdateRequest(fsUid:", qu.filesystemUid, "inodeId:", qu.InodeId, "type:", qu.quotaType, "capacity:", qu.capacityLimit, ")")
}

func NewQuotaCreateRequest(fs FileSystem, inodeId uint64, quotaType QuotaType, capacityLimit uint64) *QuotaCreateRequest {
	filesystemUid := fs.Uid
	ret := &QuotaCreateRequest{
		filesystemUid: filesystemUid,
		InodeId:       inodeId,
		quotaType:     quotaType,
		capacityLimit: capacityLimit,
	}
	if quotaType == QuotaTypeHard {
		ret.HardLimitBytes = capacityLimit
		ret.SoftLimitBytes = capacityLimit
	} else if quotaType == QuotaTypeSoft {
		ret.SoftLimitBytes = capacityLimit
		ret.HardLimitBytes = MaxQuotaSize
	}
	return ret
}

func NewQuotaUpdateRequest(fs FileSystem, inodeId uint64, quotaType QuotaType, capacityLimit uint64) *QuotaUpdateRequest {
	filesystemUid := fs.Uid
	ret := &QuotaUpdateRequest{
		filesystemUid: filesystemUid,
		InodeId:       inodeId,
		quotaType:     quotaType,
		capacityLimit: capacityLimit,
	}
	if quotaType == QuotaTypeHard {
		ret.HardLimitBytes = capacityLimit
		ret.SoftLimitBytes = capacityLimit
	} else if quotaType == QuotaTypeSoft {
		ret.SoftLimitBytes = capacityLimit
		ret.HardLimitBytes = MaxQuotaSize
	}
	return ret
}

func NewQuotaDeleteRequest(fs FileSystem, inodeId uint64) *QuotaDeleteRequest {
	return &QuotaDeleteRequest{
		filesystemUid: fs.Uid,
		InodeId:       inodeId,
	}
}

type QuotaDeleteRequest struct {
	filesystemUid uuid.UUID
	InodeId       uint64 `json:"inodeId,omitempty"`
}

func (qd *QuotaDeleteRequest) String() string {
	return fmt.Sprintln("QuotaDeleteRequest(fsUid:", qd.filesystemUid, "inodeId:", qd.InodeId, ")")
}

func (qd *QuotaDeleteRequest) getApiUrl() string {
	url, err := urlutil.URLJoin((&FileSystem{Uid: qd.filesystemUid}).GetApiUrl(), "quotas", strconv.FormatUint(qd.InodeId, 10))
	if err != nil {
		return ""
	}
	return url
}

func (qd *QuotaDeleteRequest) getRequiredFields() []string {
	return []string{"filesystemUid", "InodeId"}
}

func (qd *QuotaDeleteRequest) hasRequiredFields() bool {
	return ObjectRequestHasRequiredFields(qd)
}

func (qd *QuotaDeleteRequest) getRelatedObject() ApiObject {
	return &Quota{
		FilesystemUid: qd.filesystemUid,
		InodeId:       qd.InodeId,
	}
}

func (a *ApiClient) CreateQuota(qr *QuotaCreateRequest, q *Quota, waitForCompletion bool) error {
	f := a.Log(3, "Creating quota", qr.String(), "wait for completion:", waitForCompletion)
	if !qr.hasRequiredFields() {
		return RequestMissingParams
	}
	payload, err := json.Marshal(qr)
	if err != nil {
		return err
	}

	err = a.Put(qr.getApiUrl(), &payload, nil, q)
	if err != nil {
		return err
	}
	if waitForCompletion {
		return a.WaitForQuotaActive(q)
	}
	f()
	return nil
}

func (a *ApiClient) WaitForQuotaActive(q *Quota) error {
	glog.V(4).Infof("Waiting for quota %s@%s to become active", q.InodeId, q.FilesystemUid.String())
	f := wait.ConditionFunc(func() (bool, error) {
		return a.IsQuotaActive(q)
	})
	err := wait.Poll(5*time.Second, time.Hour*24, f)
	if err != nil {
		return err
	}
	return nil
}

func (a *ApiClient) FindQuotaByFilter(query *Quota, resultSet *[]Quota) error {
	if query.FilesystemUid == uuid.Nil {
		return RequestMissingParams
	}
	ret := &[]Quota{}
	err := a.Get(query.GetBasePath(), nil, ret)
	if err != nil {
		return err
	}
	for _, r := range *ret {
		r.FilesystemUid = query.FilesystemUid
		if r.EQ(query) {
			*resultSet = append(*resultSet, r)
		}
	}
	return nil
}

func (a *ApiClient) GetQuotaByFileSystemAndInode(fs *FileSystem, inodeId uint64) (*Quota, error) {
	if fs == nil || inodeId == 0 {
		return nil, RequestMissingParams
	}
	ret := &Quota{
		FilesystemUid: fs.Uid,
		InodeId:       inodeId,
	}
	err := a.Get(ret.GetApiUrl(), nil, ret)
	if err != nil {
		return nil, err
	}
	ret.FilesystemUid = fs.Uid
	ret.InodeId = inodeId
	return ret, nil
}

func (a *ApiClient) GetQuotaByFilter(query *Quota) (*Quota, error) {
	f := a.Log(3, "Looking for quota", query.String())
	defer f()
	rs := &[]Quota{}
	err := a.FindQuotaByFilter(query, rs)
	if err != nil {
		return nil, err
	}
	if *rs == nil || len(*rs) == 0 {
		return nil, ObjectNotFoundError
	}
	if len(*rs) > 1 {
		return nil, MultipleObjectsFoundError
	}
	result := &(*rs)[0]
	return result, nil
}

func (a *ApiClient) IsQuotaActive(query *Quota) (done bool, err error) {
	q, err := a.GetQuotaByFilter(query)
	if err != nil {
		return false, err
	}
	if q != nil {
		return q.Status == QuotaStatusActive, nil
	}
	return false, nil
}

func (a *ApiClient) UpdateQuota(r *QuotaUpdateRequest, q *Quota) error {
	f := a.Log(3, "Updating quota", r)
	defer f()
	if !r.hasRequiredFields() {
		return RequestMissingParams
	}
	var payload []byte
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = a.Put(r.getApiUrl(), &payload, nil, q)
	if err != nil {
		return err
	}
	return nil
}

func (a *ApiClient) DeleteQuota(r *QuotaDeleteRequest) error {
	f := a.Log(3, "Deleting quota", r)
	defer f()
	if !r.hasRequiredFields() {
		return RequestMissingParams
	}
	apiResponse := &ApiResponse{}
	err := a.Delete(r.getApiUrl(), nil, nil, apiResponse)
	if err != nil {
		return err
	}
	return nil
}
