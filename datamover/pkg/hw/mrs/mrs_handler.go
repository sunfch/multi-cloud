package mrshandler

import (
	"github.com/globalsign/mgo/bson"
	"time"
	"encoding/json"
	"log"
	"errors"
	"context"
	flowtype "github.com/opensds/multi-cloud/dataflow/pkg/model"
	pb "github.com/opensds/multi-cloud/datamover/proto"
	"github.com/opensds/multi-cloud/datamover/pkg/db"
	"bytes"
	"os/exec"
	"bufio"
	"io"
)


var analystDir string = "/root/gopath/src/github.com/opensds/multi-cloud/script"

var (
	JOB_STAT_TERM = "Terminated"
	JOB_STAT_STARTING = "Starting"
	JOB_STAT_RUNNING = "Running"
	JOB_STAT_COMPLETED = "Completed"
	JOB_STAT_ABNORMAL = "Abnormal"
	JOB_STAT_ERROR = "Error"
)

var (
	KEY_CLUSTER_ID  = "ClusterID"
	KEY_ENDPOINT    = "Endpoint"
	KEY_ACCESS_KEY  = "AccessKey"
	KEY_SECURITY_KEY = "SecurityKey"
	//KEY_REGION      = "Region"
	KEY_UNAME       = "Uname"
	KEY_PASSWD      = "Passwd"
	KEY_DNAME       = "Dname"
	KEY_PROJ_ID     = "ProjID"
	KEY_BUCKET      = "Bucket"
	KEY_PROG_NAME   = "ProgName"
	KEY_ARGUMENTS   = "Arguments"
)

type MrsDataAnalst struct {
	ClusterID 		string
	AccessKey		string
	SecurityKey		string
	Endpoint		string
	Uname  			string
	Passwd 			string
	Dname           string
	JobName         string
	ProjID			string
	ProgName        string
	Arguments       string
	Path            string //bucket + virt_bucket
	Token           string
	JobID           string
}

type PyResp struct {
	Err  string         `json:"err,omitempty"`
	Token string    	`json:"token,omitempty"`
	JobID string		`json:"jobid,omitempty"`
	JobStat string  	`json:"jobstat,omitempty"`
	JobProgress string  `json:"jobprogress,omitempty"`
}

func updateJob(jid bson.ObjectId, status string, stepDetails *string) error {
	j := flowtype.Job{Id:jid}
	j.Status = status
	j.StepDesc = *stepDetails
	if j.Status == flowtype.JOB_STATUS_FAILED || j.Status == flowtype.JOB_STATUS_SUCCEED {
		j.EndTime = time.Now()
	}
	var err error
	for i := 1; i <= 3; i++ {
		err := db.DbAdapter.UpdateJob(&j)
		if err == nil {
			break
		}
		if i == 3 {
			log.Printf("Update the finish status of job in database failed three times, no need to try more.")
			err = errors.New("Update job in database failed.")
		}
	}

	return err
}

func (da *MrsDataAnalst) Init(in *string) error {

	return nil
}

func createAnalysisJob(ctx context.Context, req string) (*PyResp, error) {
	log.Printf("[createAnalysisJob] Create analysis job, req:%s\n", req)
	//build request for creating analysis job

	cmd := exec.Command(
		"/root/gopath/src/github.com/opensds/multi-cloud/script/createAnalystJob.py",
         req)
	var out,errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
    err := cmd.Run()
	outs := out.String()
	errouts := errout.String()
	log.Printf("[createAnalysisJob] errout:%s\n", errouts)
	log.Printf("[createAnalysisJob] out:%s\n", outs)

    if err != nil {
		log.Printf("[createAnalysisJob] run cmd faild:%v\n", err)
		return nil, errors.New("Create analysis job failed: internal error.")
	}
    if errouts != "" {
		log.Printf("[createAnalysisJob] errout:%s\n", errouts)
    	return nil, errors.New(errouts)
	}

	rsp := PyResp{}
    err = json.Unmarshal([]byte(outs), rsp)
    if err != nil || rsp.Token == "" || rsp.JobID == "" {
		log.Printf("[createAnalysisJob] createAnalystJob.py run faild:err=%v,token:%s,jobid:%s\n",
			err, rsp.Token, rsp.JobID)
		return nil, errors.New("Create analysis job failed: internal error.")
	} else {
		log.Printf("[createAnalysisJob] createAnalystJob.py run successfully, toke:%s,jobid:%s.\n",
			rsp.Token, rsp.JobID)
	}

	return &rsp, nil
}

func updateAnalysisState(rsp *PyResp, job *flowtype.Job) {
	//job := flowtype.Job{Id:j.Id}
	if rsp.Err != "" {
		job.Status = flowtype.JOB_STATUS_FAILED
		job.StepDesc = job.StepDesc + time.Now().UTC().Format("[2006-01-02 03:04:05:06] ") +
			"Failed:" + rsp.Err
		job.EndTime = time.Now()
	} else {
		if rsp.JobStat == "Running" || rsp.JobStat == "Starting" {
			job.Status = flowtype.JOB_STATUS_RUNNING
		} else if rsp.JobStat == "Completed" {
			job.Status = flowtype.JOB_STATUS_SUCCEED
			job.StepDesc = job.StepDesc + time.Now().UTC().Format("[2006-01-02 03:04:05:06] ") +
				"succeed."
			job.EndTime = time.Now()
		} else {
			job.Status = flowtype.JOB_STATUS_FAILED
			job.StepDesc = job.StepDesc + time.Now().UTC().Format("[2006-01-02 03:04:05:06] ") +
				"analysis job failed."
			job.EndTime = time.Now()
		}
	}

	for i := 1; i <= 3; i++ {
		err := db.DbAdapter.UpdateJob(job)
		if err == nil {
			break
		}
		if i == 3 {
			log.Printf("Update the finish status of job in database failed three times, no need to try more.")
			err = errors.New("Update job in database failed.")
		}
	}
}

//run job and get result
func monitorJob(ctx context.Context, req string, j *flowtype.Job) error {
	log.Printf("[getJobFinishState] req:%v.\n", req)

	cmd := exec.Command(
		"/root/gopath/src/github.com/opensds/multi-cloud/script/getAnalystJobState.py",
		req)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
    	log.Printf("[getJobFinishState] get StdoutPipe failed, err:%v\n", err)
    	return errors.New("Analysis failed: internal error.")
	}
	cmd.Start()
    reader := bufio.NewReader(stdout)

    for {
    	out, e := reader.ReadString('\n')
    	if io.EOF == e {
			log.Println("[getJobFinishState] reader.ReadString reach the end.")
			break
		}
    	if e != nil {
			log.Printf("[getJobFinishState] reader.ReadString:e=%v\n", e)
			err = errors.New("Analysis failed: internal error.")
    		break
		}
		rsp := PyResp{}
		e = json.Unmarshal([]byte(out), rsp)
		if e != nil {
			break
		}
		updateAnalysisState(&rsp, j)
	}

	cmd.Wait()

    return err
}

func buildBasicReq(in *pb.RunJobRequest, remoteBucket string) MrsDataAnalst {
	detail := in.Asist.Details
	asist := MrsDataAnalst{}
	for _, val := range detail {
		switch val.Key {
		case KEY_CLUSTER_ID:
			asist.ClusterID = val.Value
		case KEY_ACCESS_KEY:
			asist.AccessKey = val.Value
		case KEY_SECURITY_KEY:
			asist.SecurityKey = val.Value
		case KEY_ENDPOINT:
			asist.Endpoint = val.Value
		case KEY_UNAME:
			asist.Uname = val.Value
		case KEY_PASSWD:
			asist.Passwd = val.Value
		case KEY_DNAME:
			asist.Dname = val.Value
		case KEY_PROJ_ID:
			asist.ProjID = val.Value
		case KEY_PROG_NAME:
			asist.ProgName = val.Value
		case KEY_ARGUMENTS:
			asist.Arguments = val.Value
		}
	}

	if asist.Dname == "" {
		asist.Dname = asist.Uname
	}
	asist.Path = remoteBucket + "/" + in.DestConn.BucketName
	asist.JobName = in.Id

	log.Printf("[buildBasicReq] asist:%v\n", asist)

	return asist
}

//use in.Asist.Details to build request
func buildCreateJobReq(in *pb.RunJobRequest, remoteBucket string) (string, error) {
	asist := buildBasicReq(in, remoteBucket)

	req := ""
	b, err := json.Marshal(asist)
	if err != nil {
		log.Printf("[buildCreateJobReq] marshal failed:%v\n", err)
	} else {
		req = string(b)
		log.Printf("[buildCreateJobReq] marshal successfully, req:%s\n", req)
	}

	return req, err
}

//use in.Asist.Details to build request
func buildMonitorJobReq(asistStr string, token string, jid string) (string, error) {
	asist := MrsDataAnalst{}
	err := json.Unmarshal([]byte(asistStr), asist)
	if err != nil {
		log.Printf("[buildCreateJobReq] unmarshal failed:%v\n", err)
		return "", err
	}

	asist.JobID = jid //this is the analysis job id
	asist.Token = token

	req := ""
	b, err := json.Marshal(asist)
	if err != nil {
		log.Printf("[buildCreateJobReq] marshal failed:%v\n", err)
	} else {
		req = string(b)
		log.Printf("[buildCreateJobReq] marshal successfully, req:%s\n", req)
	}

	return req, err
}

func (da *MrsDataAnalst) Handle(ctx context.Context, in *pb.RunJobRequest, j *flowtype.Job, remoteBucket string) error {
	//Create analysis job
	//createJobRsp := createAnalysisJob(module, req4Py)
	creq, err := buildCreateJobReq(in, remoteBucket)
	if err != nil {
		log.Println("buildCreateJobReq failed.")
		return errors.New("Failed:Internal error.")
	}
	createJobRsp, err := createAnalysisJob(ctx, creq)
	if err != nil {
		return err
	}
	j.StepDesc = j.StepDesc + time.Now().UTC().Format("[2006-01-02 03:04:05:06] ") +
		"create analysis job suceessfully.\n"
	updateJob(j.Id, flowtype.JOB_STATUS_RUNNING, &j.StepDesc)

	//Run job and get job finish status
	mreq, err := buildMonitorJobReq(creq, createJobRsp.Token, createJobRsp.JobID)
	if err != nil {
		log.Println("buildGetJobStatusReq failed.")
		return errors.New("Internal error.")
	}
	err = monitorJob(ctx, mreq, j)

	return err
}