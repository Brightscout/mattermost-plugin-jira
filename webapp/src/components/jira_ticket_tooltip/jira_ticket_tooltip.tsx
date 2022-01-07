import React, {PureComponent, Fragment} from "react";
import JiraAvatar from './assets/jira_avatar.png';

type Props = {
    ticketId?: string,
    ticketTitle: string,
    ticketDescription: string,
    ticketStatus: string,
    ticketStatusKey: string,
    ticketLabel: Array<string>,
    ticketAssignee: string,
    ticketAssigneePicture: string,
    ticketFixVersion?: string | null,
    ticketURI: string,
    ticketIssueTypeIconURI: string,
};

export default class TicketPopover extends PureComponent<Props> {
    truncateString(str: string, num: number) {
        if (num > str.length) {
            return str;
        } else {
            str = str.substring(0, num);
            return str + "...";
        }
    }

    //
    fixVersionLabel(fixVersion: string | undefined | null) {
        if (fixVersion) {
            return <div className="fix-version-label" style={{
                color: '#333',
                margin: '16px 0px',
                textAlign: 'left' as 'left',
                fontFamily: 'open sans',
                fontSize: '10px',
                padding: '0px 0px 2px 0px',
            }}>Fix Version: <span className="fix-version-label-value" style={{
                backgroundColor: 'rgba(63, 67, 80, 0.08)',
                padding: '1px 8px',
                fontWeight: '600',
                borderRadius: '2px',
            }}>{fixVersion}
            </span></div>;
        }
    }

    tagTicketStatus(ticketStatus: string, ticketStatusKey: string) {
        const defaultStyle = {
            fontFamily: 'open sans',
            fontStyle: 'normal',
            fontWeight: '600',
            fontSize: '12px',
            marginTop: '4px',
            padding: '4px 8px 0px 8px',
            align: 'center' as 'center',
            height: 20,
            marginBottom: '8px',
            borderRadius: '4px',
        }
        if (ticketStatusKey === "indeterminate") {
            console.log(ticketStatusKey)
            return <span style={{
                ...defaultStyle,
                color: '#FFFFFF',
                backgroundColor: '#1C58D9',
                borderRadius: '2px',
            }}>{ticketStatus}</span>
        }

        if (ticketStatusKey === "done") {
            return <span style={{
                ...defaultStyle,
                color: '#FFFFFF',
                backgroundColor: '#3DB887',

            }}>{ticketStatus}</span>
        }

        // ticketStatus == "new" or other
        return <span style={{
            ...defaultStyle,
            color: '#3F4350',
            backgroundColor: 'rgba(63, 67, 80, 0.16)',
        }}>{ticketStatus}</span>
    }

    labelList(labels: Array<string> | null | undefined) {
        if (labels !== undefined && labels !== null) {
            let totalString = 0
            let totalHide = 0;
            return (
                <Fragment>
                    <div className={'ticket-popover-label'}>
                        {labels.map(function (label: string) {
                            if (totalString >= 45 || totalString + label.length >= 45) {
                                totalHide++;
                                return null;
                            }
                            totalString += label.length + 3;
                            return <span className="jiraticket-popover-label-list" >{label}</span>;
                        })}
                    </div>
                    {
                        totalHide !== 0 ? (<div className={'jiraticket-popover-total-hide-label'}>+{totalHide}more</div>) : null
                    }

                </Fragment>
            )
        }
    }


    render() {
        const {
            ticketIssueTypeIconURI,
            ticketId,
            ticketTitle,
            ticketDescription,
            ticketFixVersion,
            ticketStatus,
            ticketStatusKey,
            ticketURI,
            ticketLabel,
            ticketAssignee,
            ticketAssigneePicture,
        } = this.props;

        return (
            <div className={'ticket-popover'}>
                <div className={'ticket-popover-header'}>
                    <div className={'ticket-popover-header-container'}>
                        <a href={ticketURI} title={'goto ticket'}>
                            <img src={JiraAvatar} width={14} height={14}
                                 alt={'jira-avatar'}
                                 className={'ticket-popover-header-avatar'}/></a>
                        <a href={ticketURI} className={'ticket-popover-keyword'}>
                            <span style={{fontSize: 12}}>{ticketId}</span>
                            <img alt={'jira-issue-icon'} width="14" height="14" src={ticketIssueTypeIconURI}/>
                        </a>
                    </div>
                </div>
                <div className={'ticket-popover-body'}>
                    <div className={'ticket-popover-title'}>
                        <a href={ticketURI}>
                            <h5>{this.truncateString(ticketTitle, 80)}</h5>
                        </a>
                        {this.tagTicketStatus(ticketStatus, ticketStatusKey)}
                    </div>
                    <div className={'ticket-popover-description'}
                         dangerouslySetInnerHTML={{__html: ticketDescription}}/>
                    <div className={'ticket-popover-labels'}>
                        {this.fixVersionLabel(ticketFixVersion)}
                        {this.labelList(ticketLabel)}
                    </div>
                </div>
                <div className={'ticket-popover-footer'}>
                    <img className={'ticket-popover-footer-assigner-profile'} src={ticketAssigneePicture} alt={'jira assigner profile'}/>
                    <span className={'ticket-popover-footer-assigner-name'}>
                            {ticketAssignee}
                        </span>
                    <span className={'ticket-popover-footer-assigner-is-assigned'}>is assigned</span>
                </div>
            </div>
        )
    }
}

