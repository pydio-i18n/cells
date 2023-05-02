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

import React, {Component} from "react"
import Pydio from 'pydio'
import PydioApi from "pydio/http/api";
import {FontIcon, CircularProgress} from 'material-ui'
import {JobsServiceApi, LogListLogRequest, ListLogRequestLogFormat} from 'cells-sdk';

const {MaterialTable} = Pydio.requireLib('components');
const {ModernTextField} = Pydio.requireLib('hoc');
const {JobsStore, moment} = Pydio.requireLib('boot');
import {debounce} from 'lodash'
import ReactJson from 'react-json-view'

const debugStorageKey = 'scheduler.logs.debug'

class TaskActivity extends Component{

    constructor(props){
        super(props);
        const serverOffset = Pydio.getInstance().Parameters.get('backend')['ServerOffset'];
        const localOffset = new Date().getTimezoneOffset() * 60
        this.state = {
            activity:[],
            loading: false,
            page:0,
            serverOffset:serverOffset+localOffset,
            timeOffset: 0,
            debug: localStorage.getItem(debugStorageKey) === 'true'
        };
    }

    toggleTimeOffset() {
        const {timeOffset, serverOffset} = this.state;
        this.setState({timeOffset:timeOffset?0:serverOffset})
    }

    componentDidMount(){
        this.loadActivity(this.props);
        this._loadDebounced = debounce((jobId) => {
            if (jobId && this.props.task && this.props.task.JobID === jobId) {
                this.loadActivity(this.props, 0, 4);
            }
        }, 500);
        JobsStore.getInstance().observe("tasks_updated", this._loadDebounced);
        const {poll} = this.props;
        if(poll){
            this._interval = window.setInterval(() => {
                if(!Pydio.getInstance().WebSocketClient.getStatus()){
                    return
                }
                this.loadActivity(this.props, (this.state && this.state.page?this.state.page:0), 4);
            }, poll);
        }
    }

    componentWillUnmount(){
        if(this._loadDebounced){
            JobsStore.getInstance().stopObserving("tasks_updated", this._loadDebounced);
        }
        if(this._interval) {
            window.clearInterval(this._interval);
        }
    }

    componentWillReceiveProps(nextProps){
        if(!this.props.task){
            this.loadActivity(nextProps);
        }
        if(nextProps.task && this.props.task && nextProps.task.ID !== this.props.task.ID){
            this.loadActivity(nextProps);
        }
    }

    loadActivity(props, page = 0, retry = 0){

        const {filter, debug} = this.state;
        const {task, poll, logTransmitter} = props;
        if(!task){
            return;
        }
        const operationId = task.JobID + '-' + task.ID.substr(0, 8);
        const api = new JobsServiceApi(PydioApi.getRestClient());

        let request = new LogListLogRequest();
        request.Query = "+OperationUuid:\"" + operationId + "\"";
        if(filter === "error") {
            request.Query += " +Level:" + filter
        } else if(filter) {
            request.Query += "+Msg:*" + filter + "*"
        }
        if(!debug) {
            request.Query += ' -Level:debug'
        }
        request.Page = page;
        request.Size = debug ? 10000 : 200;
        request.Format = ListLogRequestLogFormat.constructFromObject('JSON');
        this.setState({loading: true});
        api.listTasksLogs(request).then(response => {
            const ll = response.Logs || [];
            this.setState({activity:ll, loading: false, page: page})
            if(logTransmitter) {
                if(debug){
                    logTransmitter.set([...ll])
                } else {
                    logTransmitter.clear()
                }
            }
            if(!ll.length && retry < 3 && !poll) {
                setTimeout(() => this.loadActivity(props, page, retry + 1), 2000);
            }
        }).catch(()=>{
            this.setState({activity:[], loading: false, page: page})
            if (logTransmitter) {
                logTransmitter.clear()
            }
        });

    }

    computeTag(row) {
        const {job, descriptions} = this.props;
        const pathTag = {
            backgroundColor: '#327CA7',
            fontSize: 11,
            fontWeight: 500,
            color: 'white',
            padding: '0 8px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            borderRadius: 4,
            textAlign: 'center',
            userSelect: 'text'
        };
        let path = row.SchedulerTaskActionPath;
        let searchActions = job.Actions

        if(!path){
            return null;
        }
        if (path === 'ROOT') {
            // Special case for trigger
            return <div style={{...pathTag, backgroundColor:'white', color:'rgba(0,0,0,.87)', border: '1px solid #e0e0e0'}}>Trigger</div>
        } else if(path.indexOf('ROOT/$0$MERGE') === 0 && job.MergeAction) {
            // Special case for MergeAction
            path = path.replace('ROOT/$0$MERGE', '')
            searchActions = [job.MergeAction]
        }
        let action, specialKey;
        try{
            const obj = this.findAction(path, searchActions);
            if(obj && obj.action){
                action = obj.action
            }
            if (obj && obj.key) {
                specialKey = obj.key
                pathTag.textAlign = 'left'
                if (specialKey.indexOf('Filter') > 0 ){
                    pathTag.backgroundColor = '#F08137'
                } else {
                    pathTag.backgroundColor = '#735f3d'
                }
            }
        } catch (e) {
            //console.error(e);
        }
        if (action){
            if(action.Label){
                path = action.Label
            } else if(descriptions && descriptions[action.ID]){
                path = descriptions[action.ID].Label;
            } else if(specialKey) {
                path = specialKey
            }
        } else {
            const last = path.split('/').pop();
            const actionId = last.split('$').shift();
            if(descriptions && descriptions[actionId]){
                path = descriptions[actionId].Label;
            }
        }
        if(specialKey){
            return (
                <div style={{display:'flex'}}>
                    <span className={"mdi mdi-chevron-double-up"} style={{display: 'inline-block',marginRight: 2}}/>
                    <span style={{...pathTag, flex: 1}}>{path}</span>
                </div>
            )
        } else {
            return <div style={pathTag}>{path}</div>
        }
    }

    findAction(path, actions) {
        const parts = path.split('/');
        parts.shift();
        const actionId = [...parts].shift();
        if(actionId.indexOf('action.internal.ignored') >= 0) {
            return {}
        }
        const dols = actionId.split('$')
        const chainIndex = parseInt(dols[1]);
        const action = actions[chainIndex];
        let nextActions;
        if (dols.length > 2 && action[dols[2]]) {
            return {action: action[dols[2]], key: dols[2]}
        } else if(dols.length > 2 && dols[2] === 'MERGE' && action.MergeAction) {
            parts.shift(); // Remove current segment
            if(parts.length > 1 && action.MergeAction.ChainedActions) {
                console.log('Merge Chain', path, parts, actionId, actions, dols, action)
                return this.findAction(parts.join('/'), action.MergeAction.ChainedActions)
            } else {
                console.log('Merge Action', path, parts, actionId, actions, dols, action)
                return {action: action.MergeAction}
            }
        } else if (actionId.indexOf('$FAIL') === -1) {
            nextActions = action.ChainedActions;
        } else {
            nextActions = action.FailedFilterActions;
        }
        if(parts.length > 1) {
            // Move on step forward
            return this.findAction(parts.join('/'), nextActions);
        } else {
            return {action};
        }
    }

    computeZap(log) {
        if(!log.JsonZaps) {
            return null;
        }
        let content = {}, keyName = 'Data'
        try {
            content = JSON.parse(log.JsonZaps)
            delete(content.LogType)
            delete(content.ContentType)
            delete(content.SchedulerTaskActionTags)
            const kk = Object.keys(content)
            if(kk.length === 0){
                return null
            } else if (kk.length === 1 && content[kk[0]] instanceof Object) {
                keyName = kk
                content = content[kk[0]]
            }
        } catch (e) {
            return null;
        }
        return {content, keyName}
    }

    render(){
        const {pydio, onRequestClose} = this.props;
        const {activity, loading, page, serverOffset, timeOffset = 0, showFilters=false, filter = "", debug} = this.state;
        const cellBg = "#f5f5f5";
        const lineHeight = 32;
        const setFilter = (f) => {
            this.setState({filter:f}, ()=> this.loadActivity(this.props, 0))
        }
        const toggleDebug = () => {
            if (debug) {
                localStorage.removeItem(debugStorageKey)
            } else {
                localStorage.setItem(debugStorageKey, 'true')
            }
            this.setState({debug: !debug}, () => this.loadActivity(this.props, 0))
        }
        const tdStyle = {
            height: lineHeight,
            backgroundColor:cellBg,
            userSelect:'text',
            verticalAlign:'top',
            paddingTop: 7,
            paddingBottom: 7
        }
        const columns = [
            {name: 'SchedulerTaskActionPath', label:'', hideSmall:true, style:{...tdStyle, width:130, paddingLeft: 12, paddingRight: 0}, headerStyle:{width:130, paddingLeft: 12, paddingRight: 0}, renderCell:(row) => {
                return this.computeTag(row)
            }},
            {name:'Ts', label:pydio.MessageHash['settings.17'], style:{...tdStyle, width: 100, paddingRight: 10}, headerStyle:{width: 100, paddingRight: 10}, renderCell:(row=>{
                    const m = moment((row.Ts+timeOffset) * 1000);
                    return m.format('HH:mm:ss');
                })},
            {name:'Level', label:pydio.MessageHash['ajxp_admin.logs.level'], headerStyle:{width: 70}, style:{...tdStyle, width:70, textTransform:'uppercase', paddingRight: 0, paddingLeft: 10}, renderCell:(row)=>{
                let color;
                if(row.Level==='info') {
                    color = '#1976D0';
                } else if (row.Level === 'error') {
                    color = '#E53935';
                } else if (row.Level === 'warn') {
                    color = '#fb8c00';
                } else if (row.Level === 'debug') {
                    color = '#673AB7';
                }
                    return <span style={{color, userSelect: 'text'}}>{row.Level}</span>
            }},
            {name:'Msg', label:pydio.MessageHash['ajxp_admin.logs.message'], style:{...tdStyle, whiteSpace: 'initial'}, renderCell:(row)=> {
                    const zaps = this.computeZap(row)
                    if (zaps){
                        return <div>
                            {(row.Msg !== 'ZAPS') && <div style={{marginBottom: 8}}>{row.Msg}</div>}
                            <ReactJson collapsed={true} src={zaps.content} name={zaps.keyName}/>
                        </div>
                    } else {
                        return row.Msg
                    }
            }}
        ];
        return (
            <div style={{paddingTop: 12, paddingBottom: 10, backgroundColor:cellBg}}>
                <div style={{padding:'0 24px 10px', fontWeight:500, backgroundColor:cellBg, display:'flex', alignItems:'center'}}>
                    <div>{pydio.MessageHash['ajxp_admin.scheduler.tasks.activity.title']}</div>
                    <div style={{flex:1, textAlign:'center', fontSize: 20, display:'flex', alignItems:'center', justifyContent:'center'}}>
                        {page > 0 && <FontIcon className={"mdi mdi-chevron-left"} color={"rgba(0,0,0,.7)"} style={{cursor: 'pointer'}} onClick={()=>{this.loadActivity(this.props, page - 1)}}/>}
                        {(page > 0 || activity.length >= 200) && <span style={{fontSize: 12}}>{pydio.MessageHash[331]} {(loading?<CircularProgress size={16} thickness={1.5}/>:<span>{page + 1}</span>)}</span>}
                        {activity.length >= 200 && <FontIcon className={"mdi mdi-chevron-right"} color={"rgba(0,0,0,.7)"} style={{cursor: 'pointer'}} onClick={()=>{this.loadActivity(this.props, page + 1)}}/>}
                    </div>
                    {serverOffset !== 0 &&
                    <div style={{paddingRight: 15, cursor: "pointer"}} onClick={()=>this.toggleTimeOffset()}>
                        <FontIcon className={"mdi mdi-alarm"+(timeOffset?"-snooze":"")} color={"rgba(0,0,0,.3)"} style={{fontSize: 16}}/>
                    </div>
                    }
                    {showFilters &&
                        <div style={{zoom:.8, width:100, height:35, marginTop:-10, marginRight: 5}}>
                            <ModernTextField hintText={pydio.MessageHash['ajxp_admin.logs.3']} fullWidth={true} value={filter} onChange={(e,v)=>setFilter(v)} focusOnMount={true}/>
                        </div>
                    }
                    <div style={{paddingRight: 15, cursor: "pointer"}} onClick={() => this.setState({showFilters:!showFilters})}>
                        <FontIcon className={"mdi mdi-filter" + (showFilters ? "-remove" : "")} color={"rgba(0,0,0,.3)"} style={{fontSize: 16}}/>
                    </div>
                    <div style={{paddingRight: 15, cursor: "pointer"}} onClick={() => this.loadActivity(this.props, page)}>
                        <FontIcon className={"mdi mdi-refresh"} color={"rgba(0,0,0,.3)"} style={{fontSize: 16}}/>
                    </div>
                    <div style={{paddingRight: 15, cursor: "pointer"}} onClick={() => toggleDebug()}>
                        <FontIcon className={"mdi mdi-code-braces"} color={debug?'rgb(103,58,183)':'rgba(0,0,0,.3)'} style={{fontSize: 16}}/>
                    </div>
                    <div style={{paddingRight: 15, cursor: "pointer"}} onClick={onRequestClose}>
                        <FontIcon className={"mdi mdi-close"} color={"rgba(0,0,0,.3)"} style={{fontSize: 16}}/>
                    </div>
                </div>
                <MaterialTable
                    hideHeaders={true}
                    columns={columns}
                    data={activity}
                    showCheckboxes={false}
                    emptyStateString={loading ? <div style={{display:'flex', alignItems:'center'}}> <CircularProgress size={16} thickness={1.5}/> <span style={{flex:1, marginLeft: 5}}>{pydio.MessageHash['ajxp_admin.scheduler.tasks.activity.loading']}</span></div> : pydio.MessageHash['ajxp_admin.scheduler.tasks.activity.empty']}
                    emptyStateStyle={{backgroundColor: cellBg}}
                    computeRowStyle={(row) => {return {borderBottomColor: '#fff', height: lineHeight}}}
                />
            </div>
        )
    }

}

export {TaskActivity as default}