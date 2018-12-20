#!/usr/bin/python2.7
import sys,json,requests

reload(sys)
sys.setdefaultencoding('utf-8')

def get_token(proj_id, uname, pwd, dname):
    myurl = "https://iam.cn-north-1.myhuaweicloud.com/v3/auth/tokens"
    try:
        payload = json.dumps({"auth":{"identity":{"methods":["password"],"password":{"user":{"name":uname,"password":pwd,"domain":{"name":dname}}}},"scope":{"project":{"id":proj_id}}}})

        headers = {'Content-Type':'application/json'}
        r = requests.post(myurl, headers=headers, data=payload)
        token = ""
        if r.status_code == 200 or r.status_code == 201:
            #global token
            token = r.headers.get('X-Subject-Token')
        else:
            emsg = "Get token failed, err:" + str(r.status_code)
            sys.stderr.write(emsg)
        return token

    except Exception as e:
        emsg = "Get token failed, err:" + str(e)
        sys.stderr.write(emsg)
        return ""


def create_job(token, endpoint, proj_id, job_name, cluster_id, path, prog_name, arguments):
    #create analysis job
    headers = {'Content-Type':'application/json;charset=utf-8','X-Auth-Token':token,'X-Language':'en_us'}
    myurl = "https://" + endpoint + "/v1.1/" + proj_id + "/jobs/submit-job"
    data = {"job_type":2,"job_name":job_name,"cluster_id":"","jar_path":"","arguments":"","input":"","output":"","job_log":"","file_action":"","hql":"","hive_script_path":""}
    data['cluster_id'] = cluster_id
    data['jar_path'] = "s3a://" + path + "/program/" + prog_name
    data['arguments'] = arguments
    data['input'] = "s3a://" + path + "/data/"
    data['output'] = "s3a://" + path + "/output/"
    data['job_log'] = "s3a://" + path + "/log/"

    payload = json.dumps(data)
    job_id = ""
    r = requests.post(myurl, headers=headers, data=payload)
    if r.status_code == 200:
        json_r = r.json()
        job_id = json_r['job_execution']['id']
    else:
        #sys.stdout.write("token:" + token)
        emsg = "Create analysis job failed, err:" + str(r.status_code)
        sys.stderr.write(emsg)
        return

    return job_id


def create_and_run(req):
    try:
        reqJson = json.loads(req)
        
        uname = reqJson["Uname"]
        if ("Dname" in reqJson):
            dname = reqJson["Dname"]
        else:
            dname = uname
            
        token = get_token(reqJson["ProjID"],dname,reqJson["Passwd"],uname)
        if token == "":
            #sys.stderr.write("Get token failed.")
            return 
        
        job_id = create_job(token,reqJson["Endpoint"],reqJson["ProjID"],reqJson["JobName"],reqJson["ClusterID"],reqJson["Path"],reqJson["ProgName"],reqJson["Arguments"])
        if job_id == "":
            #sys.stderr.write("Create analysis job failed.")
            return
        else:
            ret = '{"JobID":job_id,"Token":token}'
            sys.stdout.write(job_id)
        
    except Exception as e:
        emsg = "Exception in create_and_run:" + str(e)
        sys.stderr.write(emsg)
        return


def main(argv):
    #create_and_run('{"ClusterID":"clusterid_1234156","AccessKey":"4X7JQDFTCKUNWFBRYZVC","SecurityKey":"9hr0ekZgg6vZHulEekTVfWuu1lnPFvpVAJQNHXdn","Endpoint":"mrs.cn-north-1.myhuaweicloud.com","Uname":"jack_panzer","Passwd":"dfAe423vc0D","Token":"","ProjID":"ffcae646b8824c588c3bc0f0267b3b08","Path":"obs-test-123/bket_obs","ProgName":"progname_test","Arguments":"123 46","JobName":"5c178f7a4fe60c00010e69c3"}')
    create_and_run(argv[0])

if __name__ == "__main__":
    main(sys.argv[1:])
