/*
 * Copyright 2007-2020 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
 * This file is part of Pydio.
 *
 * Pydio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */
import React from 'react'
import {Paper, FontIcon} from 'material-ui'
import PydioDataModel from 'pydio/model/data-model'
import LogTable from './LogTable';
import LogTools from './LogTools'
import Log from '../model/Log'

class LogBoard extends React.Component {

    constructor(props){
        super(props);
        this.state = {
            page: 0,
            size: 100,
            query: "",
            contentType: 'JSON',
            loading: false,
            results: 0,
        };
    }

    handleLogToolsChange(newState){
        this.setState({...newState});
    }

    componentWillReceiveProps(newProps){
        if(newProps.query !== undefined && newProps.query !== this.state.query){
            this.setState({query: newProps.query, page: 0});
        }
    }

    handleReload() {
        this.setState({z: Math.random()})
    }

    handleDecrPage() {
        this.setState({page: Math.max(this.state.page - 1, 0)})
    }

    handleIncrPage() {
        this.setState({page: this.state.page + 1})
    }

    handleTimestampContext(ts) {
        if(ts){
            const q = Log.buildTsQuery(ts, 5);
            this.setState({tmpQuery: q, focus: ts});
        } else {
            this.setState({tmpQuery: null, focus: null});
        }
    }

    handleLoadingStatusChange(status, resultsCount){
        if(this.props.onLoadingStatusChange){
            this.props.onLoadingStatusChange(status);
        } else {
            this.setState({loading: status});
        }
        this.setState({results: resultsCount});
    }

    render(){
        const {pydio, noHeader, service, disableExport, currentNode} = this.props;
        const {page, size, query, tmpQuery, focus, contentType, z, results} = this.state;
        const title = pydio.MessageHash["ajxp_admin.logs.1"];
        const buttons = (
            <LogTools
                pydio={pydio}
                service={service}
                focus={focus}
                onStateChange={this.handleLogToolsChange.bind(this)}
                disableExport={disableExport}
            />
        );

        const {body} = AdminComponents.AdminStyles();
        const blockProps = body.block.props;
        const blockStyle = body.block.container;

        const prevDisabled = page === 0;
        const nextDisabled = results < size;
        const pageSizes = [50, 100, 500, 1000];
        let paginationProps;
        if(!(prevDisabled && results < pageSizes[0]) && !focus){
            paginationProps = {
                pageSizes, prevDisabled, nextDisabled,
                onPageNext: this.handleIncrPage.bind(this),
                onPagePrev: this.handleDecrPage.bind(this),
                onPageSizeChange: (v) => {this.setState({size:v, page:0})}
            }
        }

        const mainContent = (
            <Paper {...blockProps} style={blockStyle}>
                <LogTable
                    pydio={pydio}
                    service={service || 'syslog'}
                    page={page}
                    size={size}
                    query={tmpQuery ? tmpQuery:query}
                    focus={focus}
                    contentType={contentType}
                    z={z}
                    onLoadingStatusChange={this.handleLoadingStatusChange.bind(this)}
                    onTimestampContext={this.handleTimestampContext.bind(this)}
                    {...paginationProps}
                />
            </Paper>
        );

        if(noHeader){
            return mainContent;
        } else {
            return (
                <div className="main-layout-nav-to-stack workspaces-board">
                    <div className="vertical-layout" style={{width:'100%'}}>
                        <AdminComponents.Header
                            title={title}
                            icon={currentNode.getMetadata().get('icon_class')}
                            actions={buttons}
                            reloadAction={this.handleReload.bind(this)}
                            loading={this.state.loading}
                        />
                        <div className="layout-fill">
                            {mainContent}
                            {(!service || service === 'syslog') &&
                                <div style={{padding: '0 26px', color: '#9e9e9e', fontWeight: 500}}>
                                    <u>{pydio.MessageHash['ajxp_admin.logs.sys.note']}</u> {pydio.MessageHash['ajxp_admin.logs.sys.note.content']}
                                </div>
                            }
                        </div>
                    </div>
                </div>
            );
        }

    }
}

LogBoard.propTypes = {
    dataModel:React.PropTypes.instanceOf(PydioDataModel).isRequired,
};

export {LogBoard as default}
