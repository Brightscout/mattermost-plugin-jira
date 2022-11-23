// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {ReactSelectOption} from 'types/model';

import BackendSelector, {Props as BackendSelectorProps} from '../backend_selector';

const stripHTML = (text: string) => {
    if (!text) {
        return text;
    }

    const doc = new DOMParser().parseFromString(text, 'text/html');
    return doc.body.textContent || '';
};

type Props = BackendSelectorProps & {
    searchCommentVisibilityFields: (params: {fieldValue: string}) => (
        Promise<{data: {groups: {items: {name: string}[]}}; error?: Error}>
    );
    fieldName: string;
};

export default class JiraCommentVisibilitySelector extends React.PureComponent<Props> {
    fetchInitialSelectedValues = async (): Promise<ReactSelectOption[]> => 
        (!this.props.value || (this.props.isMulti && !this.props.value.length)) ? [] : this.searchCommentVisibilityFields('');

    searchCommentVisibilityFields = (inputValue: string): Promise<ReactSelectOption[]> => {
        const params = {
            fieldValue: inputValue,
            instance_id: this.props.instanceID,
            expand: 'groups',
        };
        return this.props.searchCommentVisibilityFields(params).then(({data}) => {
            return data.groups.items.map((suggestion) => ({
                value: suggestion.name,
                label: stripHTML(suggestion.name),
            }));
        });
    };

    render = (): JSX.Element => {
        return (
            <BackendSelector
                {...this.props}
                fetchInitialSelectedValues={this.fetchInitialSelectedValues}
                search={this.searchCommentVisibilityFields}
            />
        );
    }
}
