/*
 * Copyright 2007-2017 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
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
import React, {Fragment, useState, useRef, useEffect, useCallback, useMemo} from 'react';
import Pydio from 'pydio'
import UserAvatar from '../users/avatar/UserAvatar'
import {FlatButton, IconButton, Dialog} from 'material-ui'
import {muiThemeable} from 'material-ui/styles'
const {moment} = Pydio.requireLib('boot');
import DOMUtils from 'pydio/util/dom'
import Markdown from 'react-markdown'
import Emoji from 'remark-emoji'
import GFM from 'remark-gfm'
import LinkRenderer from './LinkRenderer'
import {metaEnterToCursor} from "./chatHooks";
const {ModernTextField} = Pydio.requireLib('hoc')

/*
function useHookWithRefCallback(observer) {
    const ref = useRef(null)
    const setRef = useCallback(node => {
        if (ref.current) {
            // Make sure to cleanup any events/references added to the last instance
            console.log('cleanup')
            observer.unobserve(ref.current)
        }

        if (node) {
            // Check if a node is actually passed. Otherwise node would be null.
            // You can now do what you need to, addEventListeners, measure, etc.
            console.log('observe?')
            observer.observe(node)
        }

        // Save a reference to the node
        ref.current = node
    }, [])

    return [setRef]
}
 */

// Define OUTSIDE the rendering method
const CustomLinks = {
    a:({node, href, children, ...props}) => <LinkRenderer href={href}>{children}</LinkRenderer>
}

let Message = ({message, pydio, hideDate, sameAuthor, onDeleteMessage, edit, setEdit, onEditMessage, moreLoader, actionIconProps, muiTheme}) => {

    const [hover, setHover] = useState(false);
    const mDate = moment(parseFloat(message.Timestamp)*1000);
    const [editValue, setEditValue] = useState(message.Message)
    const [confirmDelete, setConfirmDelete] = useState(false)
    const [cursor, setCursor] = useState(-1)
    const textfieldRef = useRef(null)

    const m = (id) => pydio.MessageHash[id] || id;

    useEffect(() => {
        if(cursor > -1 && textfieldRef.current) {
            textfieldRef.current.input.refs.input.setSelectionRange(cursor, cursor)
            setCursor(-1);
        }
    }, [editValue])

    const styles = {
        date: {
            opacity: 0.53,
            textAlign: 'center',
            display: 'flex',
            margin: '5px 0',
        },
        dateLine: {
            flex: 1,
            margin: '10px 20px',
            borderBottom: '1px solid',
            opacity: 0.3
        },
        loader: {
            paddingTop: 8,
            opacity: 0.8,
            textAlign: 'center',
        },
        comment: {
            padding: '6px 16px',
            display: 'flex',
            alignItems: 'flex-start',
            backgroundColor:hover?'rgba(0,0,0,.04)':'transparent'
        },
        commentContent: {
            flex: '1',
            backgroundColor:'transparent',
            position: 'relative',
            padding: '5px 10px',
            userSelect:'text',
            webkitUserSelect:'text'
        },
        commentTitle: {
            fontSize: 16,
            fontWeight: 500,
            marginTop: -2,
            padding: '0px 0px 2px'
        },
        commentDeleteBox: {
            position: 'absolute',
            top: 5,
            right: 0,
            cursor: 'pointer',
            fontSize: 16,
            opacity:0,
            transition: DOMUtils.getBeziersTransition(),
        },
        actionBar: {
            position: 'absolute',
            top: 2,
            right: 4,
            zIndex: 2,
            cursor: 'pointer',
            fontSize: 16,
            opacity:hover||edit?1:0,
            display:'flex',
            alignItems:'center',
            color: muiTheme.palette.primary1Color,
            transition: DOMUtils.getBeziersTransition(),
        }
    };
    let authorIsLogged = false;
    if(pydio.user.id === message.Author){
        authorIsLogged = true;
    }

    let statusIndicator;
    if(message.Info && message.Info['LIVE_STATUS']) {
        statusIndicator = <div className={'dot-flashing'}/>
    }

    const avatar = (
        <div style={sameAuthor ? {visibility:'hidden'} : {paddingTop:2}}>
            <UserAvatar
                avatarSize={30}
                pydio={pydio}
                userId={message.Author}
                displayLabel={false}
                richOnHover={false}
                avatarLetters={true}
            />
        </div>
    );
    let actions = [];
    let textStyle = {...styles.commentContent};
    let deleteAction = {title:m('chat.msg.delete'), icon:'delete-outline', click: () => setConfirmDelete(true)}

    if(authorIsLogged && !edit && !statusIndicator){
        if(onEditMessage) {
            actions.push({title:m('chat.msg.edit'), icon:'pencil-outline', click: () => setEdit(true)})
        }
        actions.push(deleteAction)
    }
    let body = (
        <Markdown
            className={"chat-message-md" + (statusIndicator?' has-status-indicator':'')}
            skipHtml={true}
            urlTransform={(url) => url}
            components={CustomLinks}
            remarkPlugins={[GFM, [Emoji, {emoticon: true}]]}
        >{message.Message}</Markdown>
    )

    if (edit) {
        const save = () => {
            if(editValue && onEditMessage(message, editValue)) {
                setEdit(false)
            }
        }
        const cancel = () => {
            setEditValue(message.Message)
            setEdit(false)
        }
        if(editValue) {
            actions.push({title:m('chat.msg.save'), icon:'content-save-outline', click: save})
        } else {
            actions.push(deleteAction)
        }
        actions.push({title:m('chat.msg.revert'), icon:'undo-variant', click: cancel})
        body = (<ModernTextField
            value={editValue}
            onChange={(e,v)=>setEditValue(v)}
            inputStyle={{fontSize: 14}}
            multiLine={true}
            underlineShow={false}
            focusOnMount={true}
            fullWidth={true}
            onKeyDown={(e)=>{
                if(e.key === 'Escape') {
                    cancel()
                } else if (e.key === 'Enter') {
                    if (e.metaKey || e.ctrlKey) {
                        const {cursor, newValue} = metaEnterToCursor(e, editValue)
                        setCursor(cursor)
                        setEditValue(newValue);
                        return
                    }
                    if(editValue) {
                        save()
                    } else {
                        setConfirmDelete(true)
                    }
                }
            }}
        />);
    }

    let actionBar;
    if(actions.length) {
        actionBar = (
            <div style={styles.actionBar}>
                {actions.map(a => <IconButton {...actionIconProps} iconClassName={'mdi mdi-' + a.icon} onClick={a.click} tooltip={a.title}/> )}
            </div>)
    }

    body = <Fragment>{actionBar}{body}{statusIndicator}</Fragment>

    let containerStyle = {};
    if (sameAuthor) {
        body = <div style={textStyle}>{body}</div>
        containerStyle = {...containerStyle, marginTop: -16}
    } else {
        body = (
            <div style={textStyle}>
                <div>
                    <UserAvatar
                        labelStyle={styles.commentTitle}
                        pydio={pydio}
                        displayLabel={true}
                        displayAvatar={false}
                        userId={message.Author}
                    />
                </div>
                <div>{body}</div>
            </div>
        )
    }


    return (
        <div style={containerStyle}
             onMouseOver={()=>{setHover(true)}}
             onMouseOut={()=>{setHover(false)}}
             onContextMenu={(e) => {e.stopPropagation()}}
        >
            {authorIsLogged &&
                <Dialog
                    open={confirmDelete}
                    modal={false}
                    title={m('chat.msg.delete.confirm.title')}
                    contentStyle={{
                        background:muiTheme.dialog['containerBackground'],
                        borderRadius:muiTheme.borderRadius,
                        width:420,
                        minWidth:380,
                        maxWidth:'100%'
                    }}
                    actions={
                    [
                        <FlatButton label={m('440')} onClick={onDeleteMessage} keyboardFocused={true}/>,
                        <FlatButton label={m('441')} onClick={() => setConfirmDelete(false)}/>
                    ]
                }>
                    {m('chat.msg.delete.confirm')}
                </Dialog>
            }
            {moreLoader &&
            <div style={{...styles.loader}}>
                <FlatButton primary={true} label={m('chat.load-older')} onClick={moreLoader}/>
            </div>
            }
            {!hideDate &&
                <div style={styles.date}>
                    <span style={styles.dateLine}/>
                    <span className={"date-from"}>{mDate.fromNow()}</span>
                    <span style={styles.dateLine}/>
                </div>
            }
            <div style={styles.comment}>{avatar} {body}</div>
        </div>
    );
}
Message = muiThemeable()(Message);
export {Message as default};