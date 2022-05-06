from termcolor import colored

compName = "" 
# Get the hosts for all the teams given one test host
def get_hosts(num_of_teams, ip_format, name, os):
    fd = open(f"{compName}.csv","a")
    for i in range(1,num_of_teams+1) :
        # Replace X with team number
        ip = ip_format.replace("X",str(i))
        # Remove dots
        ip = ip.replace(".", "")
        teamNum = "team0" + str(i)
        fd.write(ip + "," + teamNum + "," + name + "," + os + "\n")

def getinput():
    global compName
    compName = input("Enter Comp name: ")
    num_of_teams=int(input(colored("Enter the number of blue teams including test team: ", "blue")))
    hostsPerTeam = int(input(colored("Enter the number of hosts per team (windows + linux + router)", "blue") + ": "))


    for i in range(hostsPerTeam):
        hostInfo = input(f"Enter the ipFormat, name, OS of host {i+1}" + colored(" [192.X.1.2, Database, Linux] ", "red") + ": ")
        hostInfo = hostInfo.replace(" ", "")
        ip, name, os = hostInfo.split(",")
        get_hosts(num_of_teams, ip, name, os)



def main():
    getinput()
    print(colored(f"Hosts written to {compName}.csv ", "green"))

if __name__ == "__main__":
    main()

