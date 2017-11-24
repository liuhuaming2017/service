package com.service.service.controller;

import org.springframework.stereotype.Component;


@Component
public class ServiceDelegate implements Service {

    public String helloworld(String name){

        // Do Some Magic Here!
        return name;
    }
}
