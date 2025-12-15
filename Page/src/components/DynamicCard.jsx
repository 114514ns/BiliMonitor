import React, {useState} from "react";
import {Button, Input, Image, Card} from "@heroui/react";
import axios from "axios";
import { Heart, MessageCircle, Share2, MoreVertical } from 'lucide-react';


function ForwardIcon() {
    return (
        <svg style={{width:'18px',height:'18px'}} xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 18 18" width="18" height="18"><path d="M9.789075 2.2956175C8.97235 1.6308450000000003 7.74999 2.212005 7.74999 3.26506L7.74999 5.3915500000000005C6.642015000000001 5.5780325 5.3073725 6.040405 4.141735000000001 7.11143C2.809155 8.335825 1.751515 10.3041 1.45716 13.404099999999998C1.409905 13.9018 1.7595399999999999 14.22505 2.105415 14.317499999999999C2.442215 14.40755 2.8807175 14.314625 3.127745 13.92915C3.9664525 12.620249999999999 4.89282 11.894575 5.765827499999999 11.50585C6.4628049999999995 11.19545 7.14528 11.093125 7.74999 11.0959L7.74999 13.235025C7.74999 14.2881 8.97235 14.869250000000001 9.789075 14.2045L15.556199999999999 9.510425000000001C16.355075 8.860149999999999 16.355075 7.640124999999999 15.556199999999999 6.989840000000001L9.789075 2.2956175zM9.165099999999999 3.0768275000000003L14.895025 7.739050000000001C15.227975 7.980475 15.235775 8.468875 14.943874999999998 8.7142L9.17615 13.416800000000002C8.979474999999999 13.562024999999998 8.75 13.4269 8.75 13.227375000000002L8.75 10.638175C8.75 10.326975000000001 8.542125 10.134725 8.2544 10.1118C7.186765 10.02955 6.1563175 10.2037 5.150895 10.69295C4.14982 11.186925 3.2102250000000003 12.096525 2.573625 13.00995C2.54981 13.046975 2.52013 13.046025 2.5211725 12.986C2.8971525 10.0573 3.9373475 8.652125 4.807025 7.85305C5.87747 6.8694775 7.213197500000001 6.444867500000001 8.2272 6.33056C8.606525 6.287802500000001 8.74805 6.0849325 8.74805 5.7032275L8.74805 3.2615475C8.74805 3.0764875000000007 8.993175 2.9321925 9.165099999999999 3.0768275000000003z" fill="currentColor"></path></svg>
    )
}

function CommentIcon() {
    return (
        <svg style={{width:'18px',height:'18px'}} xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 18 18" width="18" height="18"><path d="M1.5625 7.875C1.5625 4.595807499999999 4.220807499999999 1.9375 7.5 1.9375L10.5 1.9375C13.779175 1.9375 16.4375 4.595807499999999 16.4375 7.875C16.4375 11.0504 13.944675 13.6435 10.809275 13.80405C10.097025 14.722974999999998 8.920875 15.880675 7.267095 16.331325C6.9735075 16.4113 6.704762499999999 16.286224999999998 6.55411 16.092325C6.40789 15.904149999999998 6.3561 15.634350000000001 6.4652449999999995 15.383025C6.72879 14.776249999999997 6.776465 14.221025000000001 6.7340175 13.761800000000001C3.8167675 13.387125 1.5625 10.894475 1.5625 7.875zM7.5 2.9375C4.773095 2.9375 2.5625 5.148095 2.5625 7.875C2.5625 10.502575 4.61524 12.651075000000002 7.2041924999999996 12.8038C7.4305875 12.817174999999999 7.619625000000001 12.981200000000001 7.664724999999999 13.203475C7.772575 13.734575000000001 7.8012 14.405425000000001 7.5884275 15.148399999999999C8.748325 14.6682 9.606 13.759825 10.151275 13.016475C10.24445 12.889475 10.392050000000001 12.8138 10.54955 12.812275C13.253575 12.785725 15.4375 10.58535 15.4375 7.875C15.4375 5.148095 13.226899999999999 2.9375 10.5 2.9375L7.5 2.9375z" fill="currentColor"></path></svg>
    )
}
function  LikeIcon() {
    return (
        <svg style={{width:'18px',height:'18px'}} xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 18 18" width="18" height="18"><path d="M10.4511 2.2220125C10.218425 2.194885 10.002175 2.2953725 9.884175 2.433395C9.4264 2.9688525 9.321875 3.7501399999999996 8.978575 4.3581725C8.533574999999999 5.146395 8.1198 5.6213375 7.609775000000001 6.068507499999999C7.1751375 6.449565 6.738407499999999 6.697442499999999 6.3125 6.8050575L6.3125 14.854575C6.9198900000000005 14.868174999999999 7.572900000000001 14.876875 8.25 14.876875C9.936425 14.876875 11.367025 14.823325 12.33115 14.773699999999998C13.03235 14.737575 13.646025000000002 14.390075 13.966750000000001 13.81945C14.401900000000001 13.04535 14.9387 11.909650000000001 15.264174999999998 10.571200000000001C15.56665 9.327275 15.704699999999999 8.304325 15.766675 7.582224999999999C15.7988 7.208262500000001 15.50165 6.875019999999999 15.059999999999999 6.875019999999999L11.323274999999999 6.875019999999999C11.156575 6.875019999999999 11.000800000000002 6.791952499999999 10.907975 6.653499999999999C10.783725 6.468192500000001 10.82855 6.2670175 10.9037 6.07485C11.059 5.675084999999999 11.29355 4.9974475 11.382425000000001 4.4018275C11.470875000000001 3.80917 11.450999999999999 3.32219 11.212050000000001 2.86913C10.9571 2.3857825 10.66065 2.2464475 10.4511 2.2220125zM12.034300000000002 5.87502L15.059999999999999 5.87502C16.02035 5.87502 16.850875 6.64489 16.763 7.667825C16.697100000000002 8.435525 16.55155 9.5092 16.235825000000002 10.807500000000001C15.882625 12.259950000000002 15.3035 13.482225 14.838450000000002 14.309474999999999C14.32695 15.2194 13.377475 15.721150000000002 12.38255 15.772375C11.405125 15.822725 9.956949999999999 15.876875000000002 8.25 15.876875000000002C6.5961925 15.876875000000002 5.0846825 15.826025000000001 4.0136674999999995 15.77715C2.8370825 15.723474999999999 1.8519999999999999 14.850000000000001 1.725645 13.654824999999999C1.6404649999999998 12.849274999999999 1.5625 11.80725 1.5625 10.689375C1.5625 9.665175000000001 1.6279400000000002 8.736175 1.7045524999999997 7.998975C1.8351224999999998 6.7427075 2.9137075 5.87502 4.130655 5.87502L5.8125 5.87502C6.072015 5.87502 6.457235 5.7490675 6.9505175 5.316582499999999C7.377705000000001 4.942045 7.7193000000000005 4.5546075 8.107775 3.8665374999999997C8.492075 3.18585 8.605825 2.389785 9.124075 1.783595C9.452975 1.3988800000000001 9.99475 1.162025 10.5669 1.228745C11.16225 1.29816 11.717425 1.683875 12.09655 2.4025825000000003C12.478275 3.1262375000000002 12.474075 3.8618225 12.371500000000001 4.54938C12.302149999999997 5.0139949999999995 12.155425000000001 5.510059999999999 12.034300000000002 5.87502zM5.3125 14.82705L5.3125 6.875019999999999L4.130655 6.875019999999999C3.3792199999999997 6.875019999999999 2.77211 7.400795 2.6991975000000004 8.10235C2.6253525 8.812875 2.5625 9.70665 2.5625 10.689375C2.5625 11.762875 2.6374975 12.768475 2.7200975 13.549700000000001C2.7919925 14.229675 3.3521950000000005 14.74595 4.05924 14.778224999999999C4.4278775 14.795 4.849985 14.812050000000001 5.3125 14.82705z" fill="currentColor"></path></svg>
    )
}
function DynamicCard(props) {
    var item = props.item;
    return (
        <div onClick={() => {
            props.onClick()
        }} className={'mt-6'}>
            <Card className={`max-w-2xl mx-auto  rounded-lg shadow-sm p-4 pt-2  ${props.forward?'bg-[#F6F7F8]':''}`}  isHoverable>
                {/* Header */}
                <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-purple-200 rounded-full flex items-center justify-center overflow-hidden" >
                            <img
                                src={`${AVATAR_API}${item.TargetID || item.UID}`}
                                className="w-12 h-12"
                                style={{ borderRadius: '50%' }}
                            />
                        </div>
                        <div>
                            <h3 className="font-bold text-gray-900">{item.TargetName || item.UName}</h3>

                            <div className={'flex flex-row items-center'}>
                                <p className=" mr-2 font-light">{formatTimeAgo(item.CreatedAt || item.CreateAt)}</p>

                                {item.Title && <span className={' text-sm font-light'}>投稿了视频</span>}
                                {item.ForwardDynamic && <span className={'text-sm font-light'}>转发了动态</span>}
                            </div>

                        </div>
                    </div>

                </div>


                <div className="mb-3">
                    <p className="text-gray-800 leading-relaxed whitespace-pre-line">
                        {item.Text}
                    </p>
                </div>
                {item.ForwardDynamic && <DynamicCard item={item.ForwardDynamic} forward={true}/>}



                <div className={`${props.forward?'':'mb-4'}  ${item.Images.split(',').length%3===1?'grid grid grid-cols-3':''}`}>
                    {item.Images.split(',').map((item0,i)=>{
                        return (
                            <Image
                                key={i}
                                src={item0 + '@.avif'}
                                className={`mb-4 ${item.Images.split(',').length % 3 === 1 ? '' : ''}`}
                            />
                        );
                    })}

                </div>
                {item.Title}
                {/* Action Bar */}
                {  !props.forward&&        <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                    <button className="flex items-center gap-2 text-gray-600 hover:text-gray-800 transition-colors">
                        <ForwardIcon size={20} />
                        <span className="text-sm">{item.Forward}</span>
                    </button>

                    <button className="flex items-center gap-2 text-gray-600 hover:text-gray-800 transition-colors">
                        <CommentIcon size={20} />
                        <span className="text-sm">{item.Comments}</span>
                    </button>

                    <button
                        className={`flex items-center gap-2 transition-colors ${
                            'text-gray-600 hover:text-red-500'
                        }`}
                    >
                        <LikeIcon
                            size={20}
                            fill={ 'none'}
                            className="transition-all"
                        />
                        <span className="text-sm">{item.Like}</span>
                    </button>
                </div>}
            </Card>
        </div>

    )
}
function formatTimeAgo(timeStr) {
    // 支持 "2025-05-15T04:05:02Z" 和 "2025-05-15 12:00:00"
    const date = timeStr.includes('T')
        ? new Date(timeStr)
        : new Date(timeStr.replace(/-/g, '/'))

    const now = new Date()
    const diffMs = now - date
    const diffSec = Math.floor(diffMs / 1000)
    const diffMin = Math.floor(diffSec / 60)
    const diffHour = Math.floor(diffMin / 60)
    const diffDay = Math.floor(diffHour / 24)

    if (diffSec < 60) return '刚刚'
    if (diffMin < 60) return `${diffMin} 分钟前`
    if (diffHour < 24 && now.getDate() === date.getDate()) return `${diffHour} 小时前`

    // 昨天
    const yesterday = new Date()
    yesterday.setDate(now.getDate() - 1)
    if (
        date.getDate() === yesterday.getDate() &&
        date.getMonth() === yesterday.getMonth() &&
        date.getFullYear() === yesterday.getFullYear()
    ) {
        return `昨天 ${date.getHours().toString().padStart(2, '0')}:${date
            .getMinutes()
            .toString()
            .padStart(2, '0')}`
    }

    // 是否同一年
    const sameYear = date.getFullYear() === now.getFullYear()
    const month = date.getMonth() + 1
    const day = date.getDate()

    if (!sameYear) {
        return `${date.getFullYear()}年${month}月${day}日`
    }

    return `${month}月${day}日`
}
export default DynamicCard;