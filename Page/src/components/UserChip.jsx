import React, {useEffect, useRef} from 'react';
import {Avatar, Chip, Dropdown, DropdownItem, DropdownMenu, DropdownTrigger, Tooltip} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";

function UserChip(props) {
    return (
        <div className={'mt-2 w-full'}>
                <div >
                    <p className={'font-semibold w-full'} >{props.props.FromName}</p>
                    {(
                        <div className={'flex flex-row align-middle  mb-2  items-center'}>
                            <Avatar
                                src={`${AVATAR_API}${props.props.FromId}`}
                                onClick={() => {
                                    //toSpace(props.props.FromId);
                                    window.open("/user/" + (props.props.UID || props.props.FromId))
                                }}/>

                            {(props.props.MedalLevel || props.props.Level) ?             <Chip
                                startContent={props.props.MedalLevel || props.props.Level ?<img src={getGuardIcon(props.props.GuardLevel)}/>:<CheckIcon size={18}/> }
                                variant="faded"
                                onClick={() => {
                                    toSpace(props.props.LiverID);
                                }}
                                style={{background: props.props.UpdatedAt?(new Date().getTime() - new Date(props.props.UpdatedAt))>168*2*3600*1000?'#5762A799':getColor(props.props.Level ?? props.props.MedalLevel):getColor(props.props.Level ?? props.props.MedalLevel), color: 'white', marginLeft: '8px',marginTop:'4px'}}
                            >
                                <span>{props.props.MedalName}</span>
                                <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {props.props.MedalLevel}
                                                        </span>
                            </Chip>:<></>}
                            {props.props.Amount&& <span className='ml-auto font-bold'>{parseInt(props.props.Amount).toLocaleString()}</span>}
                        </div>
                    )}
                </div>

        </div>
    );
}

export default UserChip;