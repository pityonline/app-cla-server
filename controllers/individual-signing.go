package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigningController struct {
	baseController
}

func (this *IndividualSigningController) Prepare() {
	// sign as individual
	if this.isPostRequest() {
		this.apiPrepare(PermissionIndividualSigner)
	}
}

// @Title Post
// @Description sign as individual
// @Param	:org_cla_id	path 	string				true		"org cla id"
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @router /:org_cla_id [post]
func (this *IndividualSigningController) Post() {
	action := "sign as individual"
	sendResp := this.newFuncForSendingFailedResp(action)

	orgCLAID := this.GetString(":org_cla_id")
	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.IndividualSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}
	if ec, err := (&info).Validate(pl.Email); err != nil {
		this.sendFailedResponse(400, ec, err, action)
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		statusCode, errCode := convertDBError(err)
		this.sendFailedResponse(statusCode, errCode, err, action)
		return
	}
	if isNotIndividualCLA(orgCLA) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("invalid cla"), action)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.GetFields(); err != nil {
		statusCode, errCode := convertDBError(err)
		this.sendFailedResponse(statusCode, errCode, err, action)
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	err := (&info).Create(orgCLAID, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, true)
	if err != nil {
		statusCode, errCode := convertDBError(err)
		this.sendFailedResponse(statusCode, errCode, err, action)
		return
	}

	this.sendSuccessResp("sign successfully")
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org_repo [get]
func (this *IndividualSigningController) Check() {
	action := "check individual signing"
	org, repo := parseOrgAndRepo(this.GetString(":org_repo"))

	v, err := models.IsIndividualSigned(
		this.GetString(":platform"), org, repo, this.GetString("email"),
	)
	if err != nil {
		if statusCode, errCode := convertDBError(err); errCode != util.ErrHasNotSigned {
			this.sendFailedResponse(statusCode, errCode, err, action)
			return
		}
	}

	this.sendSuccessResp(map[string]bool{"signed": v})
}
