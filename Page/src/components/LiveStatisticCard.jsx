import LiverPage from "../pages/LiverPage";
import { Avatar, Card, CardBody, CardHeader, Image } from "@heroui/react";
import { NavLink, useNavigate } from "react-router-dom";
import React from "react";

export default function LiveStatisticCard(props) {
    const redirect = useNavigate()
    var item = props.item
    return (
        <NavLink to={`/lives/${item.ID}`} onMouseEnter={() => {
            prefetch("/api/liver/space?uid=" + item.UserID)
        }}>
            <Card className={'my-4 mx-2'} isHoverable >
                <CardHeader className="flex-col items-start">
                    <p className="text-large uppercase font-bold">{item.Title}</p>
                    <small className="text-default-500">{new Date(item.CreatedAt).toLocaleString()}</small>
                    <div className={' w-full flex flex-row'}>
                        <div className={'w-[50%] flex-row flex items-center '}>
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024"                 className=" text-black dark:text-white"
                                 fill="currentColor"
                                style={{ width: '20px', height: '20px' }}>
                                <path
                                    d="M464 512a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm200 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm-400 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm661.2-173.6c-22.6-53.7-55-101.9-96.3-143.3a444.35 444.35 0 0 0-143.3-96.3C630.6 75.7 572.2 64 512 64h-2c-60.6.3-119.3 12.3-174.5 35.9a445.35 445.35 0 0 0-142 96.5c-40.9 41.3-73 89.3-95.2 142.8-23 55.4-34.6 114.3-34.3 174.9A449.4 449.4 0 0 0 112 714v152a46 46 0 0 0 46 46h152.1A449.4 449.4 0 0 0 510 960h2.1c59.9 0 118-11.6 172.7-34.3a444.48 444.48 0 0 0 142.8-95.2c41.3-40.9 73.8-88.7 96.5-142 23.6-55.2 35.6-113.9 35.9-174.5.3-60.9-11.5-120-34.8-175.6zm-151.1 438C704 845.8 611 884 512 884h-1.7c-60.3-.3-120.2-15.3-173.1-43.5l-8.4-4.5H188V695.2l-4.5-8.4C155.3 633.9 140.3 574 140 513.7c-.4-99.7 37.7-193.3 107.6-263.8 69.8-70.5 163.1-109.5 262.8-109.9h1.7c50 0 98.5 9.7 144.2 28.9 44.6 18.7 84.6 45.6 119 80 34.3 34.3 61.3 74.4 80 119 19.4 46.2 29.1 95.2 28.9 145.8-.6 99.6-39.7 192.9-110.1 262.7z" />
                            </svg>
                            <h3 className={'mt-1 ml-0.5'}>{item.Message}</h3>
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" style={{ width: '20px', height: '20px' }}    className="ml-1 text-black dark:text-white"
                                 fill="currentColor">
                                <path
                                    d="M911.5 700.7a8 8 0 0 0-10.3-4.8L840 718.2V180c0-37.6-30.4-68-68-68H252c-37.6 0-68 30.4-68 68v538.2l-61.3-22.3c-.9-.3-1.8-.5-2.7-.5-4.4 0-8 3.6-8 8V763c0 3.3 2.1 6.3 5.3 7.5L501 910.1c7.1 2.6 14.8 2.6 21.9 0l383.8-139.5c3.2-1.2 5.3-4.2 5.3-7.5v-59.6c0-1-.2-1.9-.5-2.8zM512 837.5l-256-93.1V184h512v560.4l-256 93.1zM660.6 312h-54.5c-3 0-5.8 1.7-7.1 4.4l-84.7 168.8H511l-84.7-168.8a8 8 0 0 0-7.1-4.4h-55.7c-1.3 0-2.6.3-3.8 1-3.9 2.1-5.3 7-3.2 10.8l103.9 191.6h-57c-4.4 0-8 3.6-8 8v27.1c0 4.4 3.6 8 8 8h76v39h-76c-4.4 0-8 3.6-8 8v27.1c0 4.4 3.6 8 8 8h76V704c0 4.4 3.6 8 8 8h49.9c4.4 0 8-3.6 8-8v-63.5h76.3c4.4 0 8-3.6 8-8v-27.1c0-4.4-3.6-8-8-8h-76.3v-39h76.3c4.4 0 8-3.6 8-8v-27.1c0-4.4-3.6-8-8-8H564l103.7-191.6c.6-1.2 1-2.5 1-3.8-.1-4.3-3.7-7.9-8.1-7.9z" />
                            </svg>
                            <span className=" mt-1 ml-0.5">{item.Money}</span>
                        </div>
                        {props.showUser && <NavLink className={'w-[50%] flex justify-end'} to={'/liver/' + item.UserID}>
                            <img
                                className={'w-[40px] h-[40px] rounded-full'}
                                src={`${AVATAR_API}${item.UserID}`}
                            />
                            <span className="ml-1 my-1.5 text-default-500 cursor-pointer">{item.UserName}</span>
                        </NavLink>}
                    </div>
                </CardHeader>
                <CardBody className="overflow-visible py-2">
                    <Image
                        alt="Card background"
                        className="object-cover rounded-xl"
                        src={item.Cover}
                        isBlurred
                        isZoomed

                    />
                </CardBody>
            </Card>
        </NavLink>

    )
}
