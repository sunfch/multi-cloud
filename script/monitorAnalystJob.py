#!/usr/bin/python2.7
import sys,json,requests

reload(sys)
sys.setdefaultencoding('utf-8')

def monitor_job(req):
    try:
        reqJson = json.loads(req)

        headers = {'Content-Type':'application/json;charset=utf-8','X-Auth-Token':reqJson["Token"],'X-Language':'en_us'}
        get_job_url = "https://" + reqJson["Endpoint"] + "/v1.1/" + reqJson["ClusterID"] + "/job-exes/" + reqJson["JobID"]
        #get_job_url = "https://mrs." + reqJson["Region"] + "myhuaweicloud.com/v1.1/" + reqJson["ClusterID"] + "/job-exes/" + reqJson["JobID"]
    
        get_failed_times = 0
        loop = True
        #-1:Terminated
        #1:Starting
        #2:Running
        #3:Completed
        #4:Abnormal
        #5:Error
        job_status = -1
        job_progress = ""
        #if get job status failed with continous 3 times, we consider it is abnormal
        while get_failed_times < 3:
            get_job_r = requests.get(get_job_url, headers=headers)
            if get_job_r.status_code == 200:
                get_failed_times = 0
                json_r =  get_job_r.json()
                job_status = json_r['job_execution']['job_state']
                job_progress = json_r['job_execution']['progress']
                if job_status != 1 and job_status != 2:
                    loop = False
                else:
                    sys.stdout.write(json.dumps({"status":dict[str(job_status)],"progress":job_progress}))
                    sys.stdout.flush()
            else:
                get_failed_times += 1
                #print(" failed once, return code of REST request is:%d" % r.status_code)
            time.sleep(3)
    
        dict = {'-1':'Terminated', '1':'Starting', '2':'Running', '3':'Completed', '4':'Abnormal', '5':'Error'}
        if get_failed_times >= 3:
            sys.stdout.write(json.dumps({"err":"Get job failed more than three times, maybe network unstable."})
        else:
            sys.stdout.write(json.dumps({"status":dict[str(job_status)],"progress":job_progress}))

        return

    except Exception as e:
        emsg = "Exception in monitor_job:" + str(e)
        sys.stdout.write(json.dumps({"err":emsg}))
        return 



def main(argv):
    monitor_job(argv[0])

if __name__ == "__main__":
    main(sys.argv[1:])
