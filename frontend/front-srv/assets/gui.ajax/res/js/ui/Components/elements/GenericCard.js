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
import React from 'react'
import Pydio from 'pydio'
import {Paper, IconButton, FlatButton, RaisedButton, FontIcon, IconMenu, FloatingActionButton} from 'material-ui'
import {muiThemeable} from 'material-ui/styles';
const {PlaceHolder, PhRoundShape, PhTextRow, AdditionalIcons} = Pydio.requireLib('hoc')

const globalStyles = {
    globalLeftMargin : 64,
};

const Mui3CardLine = muiThemeable() (({legend, data, legendStyle, dataStyle, placeHolder, placeHolderReady, muiTheme}) => {

    const style = {
        legend: {
            fontSize: 14,
            color: muiTheme.palette.mui3['on-surface'],
            fontWeight: 500,
            ...legendStyle,
        },
        data: {
            fontSize: 12,
            color: muiTheme.palette.mui3['on-surface-variant'],
            paddingRight: 6,
            overflow:'hidden',
            textOverflow:'ellipsis',
            ...dataStyle,
        }
    };

    return (
        <div style={{marginBottom: 20}}>
            <div style={style.legend}>{legend}</div>
            <div style={style.data}>{data}</div>
        </div>
    )

})

class GenericLine extends React.Component{
    render(){
        const {iconClassName, legend, data, dataStyle, legendStyle,
            iconStyle, placeHolder, placeHolderReady, muiTheme} = this.props;
        const style = {
            icon: {
                margin:'16px 20px 0',
                ...iconStyle,
            },
            legend: {
                fontSize: 12,
                color: muiTheme.palette.mui3['on-surface'],
                fontWeight: 500,
                ...legendStyle,
            },
            data: {
                fontSize: 14,
                paddingRight: 6,
                overflow:'hidden',
                textOverflow:'ellipsis',
                ...dataStyle,
            }
        };
        const contents = (
            <div style={{display:'flex', marginBottom: 8, overflow:'hidden', ...this.props.style}}>
                <div style={{width: globalStyles.globalLeftMargin}}>
                    <FontIcon color={muiTheme.palette.mui3['secondary']||'#aaaaaa'} className={iconClassName} style={style.icon}/>
                </div>
                <div style={{flex: 1}}>
                    <div style={style.legend}>{legend}</div>
                    <div style={style.data}>{data}</div>
                </div>
            </div>
        );
        if (placeHolder) {
            const linePH = (
                <div style={{display:'flex', marginBottom: 16, overflow:'hidden', ...this.props.style}}>
                    <div style={{width: globalStyles.globalLeftMargin}}>
                        <PhRoundShape style={{width:35,height:35,margin:'10px 15px 0'}}/>
                    </div>
                    <div style={{flex: 1}}>
                        <div style={{...style.legend,maxWidth:100}}><PhTextRow/></div>
                        <div style={{...style.data, marginRight:24}}><PhTextRow style={{height:'1.3em', marginTop:'0.4em'}}/></div>
                    </div>
                </div>
            );
            return (
                <PlaceHolder ready={placeHolderReady} showLoadingAnimation customPlaceholder={linePH}>
                    {contents}
                </PlaceHolder>
            );
        }
        return contents;
    }
}
GenericLine = muiThemeable()(GenericLine);

class GenericCard extends React.Component{

    render(){

        const {title, onDismissAction, onEditAction, onDeleteAction, otherActions, moreMenuItems,
            children, muiTheme, style, headerSmall, editTooltip, deleteTooltip, mui3 = false, topLeftAvatar} = this.props;

        const headerBg = muiTheme.palette.mui3['secondary-container'];
        const headerColor = muiTheme.palette.mui3['on-secondary-container'];
        const buttonColor = muiTheme.palette.mui3['primary']

        let styles = {
            headerHeight: 'auto',
            buttonBarHeight: 60,
            headerBg,
            headerColor,
            buttonBar:{
                display:'flex',
                height: 60
            },
            fabTop: 80,
            button: {
                style:{
                    height: 40, width: 40, padding: 8
                },
                iconStyle:{color:buttonColor},
            },
            title: {
                paddingLeft: onEditAction?globalStyles.globalLeftMargin:20,
                fontSize: 20,
                lineHeight: '26px',
                paddingBottom: 16
            }
        };
        if(mui3) {
            styles.headerBg = 'transparent'
            styles.childrenContainer = {padding:'0 20px'};
            styles.buttonBar = {
                padding:20,
                display:'flex',
                alignItems:'center',
            }
            styles.title= {
                fontSize: 22,
                lineHeight: '26px',
                padding: '0px 20px 16px'
            }
        }
        if (headerSmall) {
            styles = {
                headerHeight: 'auto',
                headerBg,
                headerColor,
                buttonBar: {
                    display: 'flex',
                    alignItems:'center',
                    height: 42,
                    padding: '0 7px 0 16px'
                },
                fabTop: 60,
                button: {
                    style:{width:38, height: 38, padding: 9},
                    iconStyle:{color:buttonColor, fontSize: 18}
                }
            }
        }

        const {DeleteOutline} = AdditionalIcons;

        return (
            <div style={{width: '100%', position:'relative', overflowX: 'hidden', ...style}}>
                <Paper zDepth={0} style={{backgroundColor:styles.headerBg, color: styles.headerColor, height: styles.headerHeight, borderRadius: '2px 2px 0 0'}}>
                    <div style={styles.buttonBar}>
                        {topLeftAvatar}
                        {headerSmall && <span style={{flex: 1, fontSize: 14, fontWeight:500}}>{title}</span>}
                        {!headerSmall && <span style={{flex: 1}}/>}
                        {otherActions}
                        {onEditAction && headerSmall &&
                            <IconButton style={styles.button.style} iconStyle={styles.button.iconStyle} iconClassName={"mdi mdi-pencil"} onClick={onEditAction} tooltip={editTooltip} tooltipPosition={"bottom-left"}/>
                        }
                        {onDeleteAction && headerSmall &&
                            <IconButton style={{...styles.button.style, padding: 7}} iconStyle={styles.button.iconStyle} onClick={onDeleteAction} tooltip={deleteTooltip} tooltipPosition={"bottom-left"}><DeleteOutline/></IconButton>
                        }
                        {moreMenuItems && moreMenuItems.length > 0 &&
                            <IconMenu
                                anchorOrigin={{vertical:'top', horizontal:headerSmall?'right':'left'}}
                                targetOrigin={{vertical:'top', horizontal:headerSmall?'right':'left'}}
                                iconButtonElement={<IconButton style={styles.button.style} iconStyle={styles.button.iconStyle} iconClassName={"mdi mdi-dots-vertical"}/>}
                            >{moreMenuItems}</IconMenu>
                        }
                        {onDismissAction &&
                            <IconButton  style={{...styles.button.style, backgroundColor:muiTheme.palette.mui3['surface-variant'], borderRadius:'50%'}} iconStyle={styles.button.iconStyle} iconClassName={"mdi mdi-close"} onClick={onDismissAction}/>
                        }
                    </div>
                    {!headerSmall && <div style={styles.title}>{title}</div>}
                </Paper>
                <div style={{paddingTop: 12, paddingBottom: 8, position:'relative', ...styles.childrenContainer}}>
                    {!mui3 && onEditAction && !headerSmall &&
                        <FloatingActionButton onClick={onEditAction} backgroundColor={muiTheme.palette.mui3['tertiary']} mini={true} style={{position:'absolute', top:-20, left: 10}}>
                            <FontIcon className={"mdi mdi-pencil"} style={{color:muiTheme.palette.mui3['on-tertiary']}} />
                        </FloatingActionButton>
                    }
                    {children}
                </div>
                {(onEditAction || onDeleteAction) && !headerSmall && mui3 &&
                    <div style={{padding:'12px 0', margin:'0 20px', display:'flex', borderTop: '1px solid ' + muiTheme.palette.mui3['outline-variant']}}>
                        <span style={{flex: 1}}/>
                        {onDeleteAction && <FlatButton label={deleteTooltip} onClick={()=>onDeleteAction()}/>}
                        {onEditAction && <RaisedButton style={{marginLeft: 5}} label={editTooltip} onClick={()=>onEditAction()}/>}
                    </div>
                }
            </div>
        );
    }

}

GenericCard = muiThemeable()(GenericCard);
export {GenericCard, GenericLine, Mui3CardLine}