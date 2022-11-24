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

export default function JiraCommentVisibilitySelector(props: Props) {
    const fetchInitialSelectedValues = async (): Promise<ReactSelectOption[]> => 
        (!props.value || (props.isMulti && !props.value.length)) ? [] : searchCommentVisibilityFields('');

        const searchCommentVisibilityFields = async (inputValue: string): Promise<ReactSelectOption[]> => {
        const params = {
            fieldValue: inputValue,
            instance_id: props.instanceID,
            expand: 'groups',
        };
        return props.searchCommentVisibilityFields(params).then(({data}) => {
            return data.groups.items.map((suggestion) => ({
                value: suggestion.name,
                label: stripHTML(suggestion.name),
            }));
        });
    };

    return (
        <BackendSelector
            {...props}
            fetchInitialSelectedValues={fetchInitialSelectedValues}
            search={searchCommentVisibilityFields}
        />
    );
}
