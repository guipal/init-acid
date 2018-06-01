package incubator

var (
	//STATIC variables --------------------------------------------------------
	BASEPROJECT = "JEN"
	USEPROJECT  = "JIC"
	BASEBRANCH  = "master"

	REPOMASTERIMAGE = "jenkins-master-image-deploy"
	REPOCONFIGTPL   = "jenkins-master-config-%v"

	SecurityMsg        = "You are up to make some major changes which you might NOT revert. For security purposes please use the flag -f."
	SomethingWentWrong = "Something went wrong, ERROR: %s."
	FatalBitbucketMsg  = "Bitbucket API call failed"
	ForceTagMsg        = "Tag already exist, use -f to force the action."

	SyncStatus int
)
