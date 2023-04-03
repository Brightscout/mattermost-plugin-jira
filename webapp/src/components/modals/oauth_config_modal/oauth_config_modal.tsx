// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {PureComponent} from 'react';
import {Modal} from 'react-bootstrap';

import OAuthConfigModalForm from './oauth_config_modal_form';
import {Props} from './props';

export default function OAuthConfigModal(props: Props) {
    const {visible} = props;
    let content;
    if (visible) {
        content = (
            <OAuthConfigModalForm {...props}/>
        );
    }

    return (
        <Modal
            dialogClassName='modal--scroll'
            show={visible}
            onHide={props.closeModal}
            onExited={props.closeModal}
            bsSize='large'
            backdrop='static'
        >
            <Modal.Header closeButton={true}>
                <Modal.Title>
                    {'Configure your Jira Cloud OAuth 2.0'}
                </Modal.Title>
            </Modal.Header>
            {content}
        </Modal>
    );
}
